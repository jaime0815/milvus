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

	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/mq/msgstream/mqwrapper"
)

type kafkaProducer struct {
	p         *kafka.Producer
	topic     string
	closeOnce sync.Once
	client    *kafkaClient
}

func NewKafkaProducer(p *kafka.Producer, topic string, client *kafkaClient) *kafkaProducer {
	kp := &kafkaProducer{p: p, client: client, topic: topic}
	return kp
}

func (kp *kafkaProducer) Topic() string {
	return kp.topic
}

func (kp *kafkaProducer) SendBatch(ctx context.Context, messages []*mqwrapper.ProducerMessage) ([]mqwrapper.MessageID, error) {
	ch := kp.client.getOrCreateDeliveryChan(kp.topic)
	for _, message := range messages {
		kp.p.ProduceChannel() <- &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &kp.topic, Partition: mqwrapper.DefaultPartitionIdx},
			Value:          message.Payload,
		}
	}

	retSize := len(messages)
	msgIDs := make([]mqwrapper.MessageID, 0, retSize)
	errs := make(errorutil.ErrorList, 0, retSize)

	for ret := range ch {
		msgIDs = append(msgIDs, &kafkaID{messageID: ret.msgOffset})
		if ret.err != nil {
			errs = append(errs, ret.err)
		}

		if len(msgIDs) == retSize {
			if len(errs) != 0 {
				return msgIDs, errors.New(errs.Error())
			}
			return msgIDs, nil
		}
	}

	return nil, fmt.Errorf("kafka send batch failed, topic:%s", kp.topic)
}

func (kp *kafkaProducer) Send(ctx context.Context, message *mqwrapper.ProducerMessage) (mqwrapper.MessageID, error) {
	ch := kp.client.getOrCreateDeliveryChan(kp.topic)

	kp.p.ProduceChannel() <- &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &kp.topic, Partition: mqwrapper.DefaultPartitionIdx},
		Value:          message.Payload,
	}

	ret := <-ch
	return &kafkaID{messageID: ret.msgOffset}, ret.err
}

func (kp *kafkaProducer) Close() {
	kp.closeOnce.Do(func() {
		start := time.Now()
		//flush in-flight msg within queue.
		kp.p.Flush(10000)

		kp.client.removeDeliveryChan(kp.topic)

		cost := time.Since(start).Milliseconds()
		if cost > 500 {
			log.Debug("kafka producer is closed", zap.Any("topic", kp.topic), zap.Int64("time cost(ms)", cost))
		}
	})
}
