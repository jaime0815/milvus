package kafka

import (
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/mq/msgstream/mqwrapper"
	"go.uber.org/zap"
)

type kafkaClient struct {
	// more configs you can see https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md
	// how to decrease message latency https://docs.confluent.io/5.5.1/clients/librdkafka/md_INTRODUCTION.html
	basicConfig kafka.ConfigMap
}

func NewKafkaClientInstance(address string) *kafkaClient {
	config := kafka.ConfigMap{
		"bootstrap.servers":   address,
		"socket.timeout.ms":   300000,
		"socket.max.fails":    3,
		"api.version.request": true,
		//"debug":               "broker,topic,msg,consumer,queue,fetch,cgrp",
	}

	return &kafkaClient{basicConfig: config}
}

func cloneKafkaConfig(config kafka.ConfigMap) *kafka.ConfigMap {
	newConfig := make(kafka.ConfigMap)
	for k, v := range config {
		newConfig[k] = v
	}
	return &newConfig
}

func (kc *kafkaClient) newProducerConfig() *kafka.ConfigMap {
	newConf := cloneKafkaConfig(kc.basicConfig)
	// default max message size 5M
	newConf.SetKey("message.max.bytes", 10485760)
	newConf.SetKey("compression.codec", "zstd")
	newConf.SetKey("queue.buffering.max.ms", 5)
	return newConf
}

func (kc *kafkaClient) newConsumerConfig(group string, offset mqwrapper.SubscriptionInitialPosition) *kafka.ConfigMap {
	newConf := cloneKafkaConfig(kc.basicConfig)

	if offset == mqwrapper.SubscriptionPositionEarliest {
		newConf.SetKey("auto.offset.reset", "earliest")
	} else {
		newConf.SetKey("auto.offset.reset", "earliest")
		//newConf.SetKey("auto.offset.reset", "latest")
	}

	newConf.SetKey("session.timeout.ms", 180000)
	newConf.SetKey("group.id", group)
	newConf.SetKey("enable.auto.commit", false)

	//Kafka default will not create topics if consumer's the topics don't exist.
	//In order to compatible with other MQ, we need to enable the following configuration,
	//meanwhile, some implementation also try to consume a non-exist topic, such as dataCoordTimeTick.
	newConf.SetKey("allow.auto.create.topics", true)

	newConf.SetKey("go.events.channel.enable", true)

	newConf.SetKey("heartbeat.interval.ms", 30000)
	newConf.SetKey("fetch.wait.max.ms", 10)
	return newConf
}

func (kc *kafkaClient) CreateProducer(options mqwrapper.ProducerOptions) (mqwrapper.Producer, error) {
	config := kc.newProducerConfig()
	pp, err := kafka.NewProducer(config)
	if err != nil {
		log.Error("kafka create sync producer , error", zap.Error(err))
		return nil, err
	}

	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()
	//client, err := kafka.NewAdminClientFromProducer(pp)
	//defer client.Close()
	//if err != nil {
	//	return nil, err
	//}
	//
	//newTopic := []kafka.TopicSpecification{{Topic: options.Topic, NumPartitions: 1}}
	//ret, err := client.CreateTopics(ctx, newTopic, kafka.SetAdminValidateOnly(true))
	//if err != nil {
	//	return nil, err
	//}
	//
	//if ret[0].Error.Code() != kafka.ErrNoError {
	//	return nil, ret[0].Error
	//}

	deliveryChan := make(chan kafka.Event, 128)
	producer := &kafkaProducer{p: pp, deliveryChan: deliveryChan, topic: options.Topic}
	return producer, nil
}

func (kc *kafkaClient) Subscribe(options mqwrapper.ConsumerOptions) (mqwrapper.Consumer, error) {
	config := kc.newConsumerConfig(options.SubscriptionName, options.SubscriptionInitialPosition)
	consumer := newKafkaConsumer(config, options.Topic, options.SubscriptionName)
	return consumer, nil
}

func (kc *kafkaClient) EarliestMessageID() mqwrapper.MessageID {
	return &kafkaID{messageID: int64(kafka.OffsetBeginning)}
}

func (kc *kafkaClient) StringToMsgID(id string) (mqwrapper.MessageID, error) {
	offset, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	return &kafkaID{messageID: offset}, nil
}

func (kc *kafkaClient) BytesToMsgID(id []byte) (mqwrapper.MessageID, error) {
	offset := DeserializeKafkaID(id)
	return &kafkaID{messageID: offset}, nil
}

func (kc *kafkaClient) Close() {
}
