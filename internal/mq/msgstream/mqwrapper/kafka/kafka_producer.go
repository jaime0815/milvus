package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/milvus-io/milvus/internal/log"
	"go.uber.org/zap"

	"github.com/confluentinc/confluent-kafka-go/kafka"
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

func (kp *kafkaProducer) Send(ctx context.Context, message *mqwrapper.ProducerMessage) (mqwrapper.MessageID, error) {
	err := kp.p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &kp.topic, Partition: mqwrapper.DefaultPartitionIdx},
		Value:          message.Payload,
	}, kp.deliveryChan)

	if err != nil {
		return nil, err
	}

	e := <-kp.deliveryChan
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		return nil, m.TopicPartition.Error
	}

	fmt.Println("==========send", m)
	kp.p.Flush(1000)
	return &kafkaID{messageID: int64(m.TopicPartition.Offset)}, nil
}

func (kp *kafkaProducer) Close() {
	log.Debug("starting close kafka producer", zap.Any("topic", kp.topic))
	kp.closeOnce.Do(func() {
		go kp.p.Close()
		log.Debug("starting close kafka producer", zap.Any("topic", kp.topic))
	})
}
