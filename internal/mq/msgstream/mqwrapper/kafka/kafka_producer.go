package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/milvus-io/milvus/internal/util/errorutil"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"

	"github.com/milvus-io/milvus/internal/common"
	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/mq/msgstream/mqwrapper"
)

type kafkaProducer struct {
	p            *kafka.Producer
	topic        string
	deliveryChan chan kafka.Event
	closeOnce    sync.Once
}

func (kp *kafkaProducer) Topic() string {
	return kp.topic
}

func (kp *kafkaProducer) SendBatch(ctx context.Context, messages []*mqwrapper.ProducerMessage) ([]mqwrapper.MessageID, error) {
	retCount := len(messages)
	errs := make(errorutil.ErrorList, 0, retCount)

	for _, message := range messages {
		err := kp.p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &kp.topic, Partition: mqwrapper.DefaultPartitionIdx},
			Value:          message.Payload,
		}, kp.deliveryChan)

		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return nil, errors.New(errs.Error())
	}

	rets, err := kp.collectRets(retCount)
	if err != nil {
		return nil, err
	}

	return rets, nil
}

func (kp *kafkaProducer) Send(ctx context.Context, message *mqwrapper.ProducerMessage) (mqwrapper.MessageID, error) {
	err := kp.p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &kp.topic, Partition: mqwrapper.DefaultPartitionIdx},
		Value:          message.Payload,
	}, kp.deliveryChan)

	if err != nil {
		return nil, err
	}

	rets, err := kp.collectRets(1)
	if err != nil {
		return nil, err
	}

	return rets[0], nil
}

func (kp *kafkaProducer) collectRets(count int) ([]mqwrapper.MessageID, error) {
	i := 0
	msgIDs := make([]mqwrapper.MessageID, count)
	for i != count {
		e, ok := <-kp.deliveryChan
		if !ok {
			log.Error("kafka produce message fail because of delivery chan is closed", zap.String("topic", kp.topic))
			return nil, common.NewIgnorableError(fmt.Errorf("delivery chan of kafka producer is closed"))
		}

		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			return nil, m.TopicPartition.Error
		}

		msgIDs[i] = &kafkaID{messageID: int64(m.TopicPartition.Offset)}
		i++
	}
	return msgIDs, nil
}

func (kp *kafkaProducer) Close() {
	kp.closeOnce.Do(func() {
		start := time.Now()
		//flush in-flight msg within queue.
		kp.p.Flush(10000)

		close(kp.deliveryChan)

		cost := time.Since(start).Milliseconds()
		if cost > 500 {
			log.Debug("kafka producer is closed", zap.Any("topic", kp.topic), zap.Int64("time cost(ms)", cost))
		}
	})
}
