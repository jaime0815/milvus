package kafka

import (
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/mq/msgstream/mqwrapper"
	"go.uber.org/zap"
)

type Consumer struct {
	c          *kafka.Consumer
	config     *kafka.ConfigMap
	msgChannel chan mqwrapper.Message
	isStarted  bool
	skipMsg    bool
	topic      string
	groupID    string
	closeCh    chan struct{}
	chanOnce   sync.Once
	closeOnce  sync.Once
	wg         sync.WaitGroup
}

func newKafkaConsumer(config *kafka.ConfigMap, topic string, groupID string) *Consumer {
	closeCh := make(chan struct{})
	msgChannel := make(chan mqwrapper.Message, 1)

	kafkaConsumer := &Consumer{
		config:     config,
		msgChannel: msgChannel,
		topic:      topic,
		groupID:    groupID,
		closeCh:    closeCh,
	}

	kafkaConsumer.createKafkaConsumer()
	return kafkaConsumer
}

func (kc *Consumer) createKafkaConsumer() error {
	var err error
	kc.c, err = kafka.NewConsumer(kc.config)
	if err != nil {
		log.Fatal("create kafka consumer failed", zap.String("topic", kc.topic), zap.Error(err))
		return err
	}

	err = kc.c.Subscribe(kc.topic, nil)
	if err != nil {
		log.Fatal("kafka consumer subscribe failed", zap.String("topic", kc.topic), zap.Error(err))
		return err
	}

	return nil
}

func (kc *Consumer) readMessage() *kafka.Message {
	msg, err := kc.c.ReadMessage(100 * time.Millisecond)

	if err != nil {
		// Kafka will not support read message before creating topic.
		// In order to compatible with other mq, we need hold the following error.
		errCode := err.(kafka.Error).Code()
		if errCode == kafka.ErrUnknownTopicOrPart || errCode == kafka.ErrTimedOut {
			return nil
		}

		log.Error("read msg failed", zap.Any("topic", kc.topic), zap.String("groupID", kc.groupID), zap.Error(err))
		return nil
	}
	return msg
}

func (kc *Consumer) startReceiveMsgTask() {
	// clean data within channel
	for len(kc.msgChannel) > 0 {
		<-kc.msgChannel
	}

	if kc.isStarted {
		return
	}

	go func() {
		if kc.skipMsg {
			msg := kc.readMessage()
			if msg != nil {
				kc.skipMsg = false
			}
		}

		msg := kc.readMessage()

		select {
		case <-kc.closeCh:
			close(kc.msgChannel)
			return
		default:
			if msg != nil {
				kc.msgChannel <- &kafkaMessage{msg: msg}
			}
		}
	}()

	kc.isStarted = true
}

// internalClose close kafka consumer
func (kc *Consumer) internalClose() error {
	if err := kc.c.Unassign(); err != nil {
		return err
	}

	if err := kc.c.Unsubscribe(); err != nil {
		return err
	}

	if err := kc.c.Close(); err != nil {
		return err
	}

	return nil
}

func (kc *Consumer) Subscription() string {
	return kc.groupID
}

// Chan provides a channel to read consumed message.
// There are some illustrations need to clarify,
// 1.confluent-kafka-go recommend us to use function-based consumer,
// channel-based consumer API had already deprecated, see more details
// https://github.com/confluentinc/confluent-kafka-go.
//
// 2.This API of other MQ return channel directly, but it depends on
// readMessage firstly which means it be always triggered within select-case
// invocation. However, it still works well, because it covers all messages
// consume situation: start from the earliest or latest position to keep consume;
// start from a seek position to specified end position.
func (kc *Consumer) Chan() <-chan mqwrapper.Message {
	kc.chanOnce.Do(func() {
		kc.startReceiveMsgTask()
	})
	return kc.msgChannel
}

func (kc *Consumer) Seek(id mqwrapper.MessageID, inclusive bool) error {
	if err := kc.createKafkaConsumer(); err != nil {
		return err
	}

	offset := kafka.Offset(id.(*kafkaID).messageID)
	log.Debug("kafka consumer seek ", zap.String("topic name", kc.topic),
		zap.Any("Msg offset", offset), zap.Bool("inclusive", inclusive))

	//There is need to invoke Unassign before Assign or seek twice will fail
	//on the same topic and partition.
	if err := kc.c.Unassign(); err != nil {
		log.Error("kafka consumer unassign failed ", zap.String("topic name", kc.topic), zap.Any("Msg offset", offset), zap.Error(err))
		return err
	}

	err := kc.c.Assign([]kafka.TopicPartition{{Topic: &kc.topic, Partition: mqwrapper.DefaultPartitionIdx, Offset: offset}})
	if err != nil {
		log.Error("kafka consumer assign failed ", zap.String("topic name", kc.topic), zap.Any("Msg offset", offset), zap.Error(err))
		return err
	}

	// If seek timeout is not 0 the call twice will return error isStarted RD_KAFKA_RESP_ERR__STATE.
	// if the timeout is 0 it will initiate the seek  but return immediately without any error reporting
	kc.skipMsg = !inclusive
	err = kc.c.Seek(kafka.TopicPartition{
		Topic:     &kc.topic,
		Partition: mqwrapper.DefaultPartitionIdx,
		Offset:    offset}, 1000)
	if err != nil {
		return err
	}

	kc.startReceiveMsgTask()
	return nil
}

func (kc *Consumer) Ack(message mqwrapper.Message) {
	//offset := kafka.Offset(message.ID().(*kafkaID).messageID)
	//partitions := make([]kafka.TopicPartition, 1)
	//partitions[0] = kafka.TopicPartition{Topic: &kc.topic, Partition: mqwrapper.DefaultPartitionIdx, Offset: offset}
	//kc.c.CommitOffsets(partitions)
	kc.c.Commit()
}

func (kc *Consumer) GetLatestMsgID() (mqwrapper.MessageID, error) {
	low, high, err := kc.c.QueryWatermarkOffsets(kc.topic, mqwrapper.DefaultPartitionIdx, -1)
	if err != nil {
		return nil, err
	}

	// Current high value is next offset of the latest message ID, in order to keep
	// semantics consistency with the latest message ID, the high value need to move forward.
	if high > 0 {
		high = high - 1
	}

	log.Debug("get latest msg ID ", zap.Any("topic", kc.topic), zap.Int64("oldest offset", low), zap.Int64("latest offset", high))
	return &kafkaID{messageID: high}, nil
}

func (kc *Consumer) Close() {
	kc.closeOnce.Do(func() {
		log.Debug("starting close kafka consumer", zap.Any("topic", kc.topic), zap.String("groupID", kc.groupID))
		start := time.Now()
		close(kc.closeCh)
		//kc.c.Unsubscribe()
		kc.c.Close()
		log.Debug("close kafka consumer finished ", zap.Any("topic", kc.topic), zap.String("groupID", kc.groupID), zap.Any("token time", zap.Int64("time cost(ms)", time.Since(start).Milliseconds())))
	})
}
