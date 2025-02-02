package sarama_test

import (
	"log"

	saramatrace "github.com/adityayuga/signalfx-go-tracing/contrib/Shopify/sarama"
	"github.com/adityayuga/signalfx-go-tracing/ddtrace/tracer"
	sarama "gopkg.in/Shopify/sarama.v1"
)

func Example_asyncProducer() {
	cfg := sarama.NewConfig()

	producer, err := sarama.NewAsyncProducer([]string{"localhost:9092"}, cfg)
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	producer = saramatrace.WrapAsyncProducer(cfg, producer)

	msg := &sarama.ProducerMessage{
		Topic: "some-topic",
		Value: sarama.StringEncoder("Hello World"),
	}
	producer.Input() <- msg
}

func Example_syncProducer() {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, cfg)
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	producer = saramatrace.WrapSyncProducer(cfg, producer)

	msg := &sarama.ProducerMessage{
		Topic: "some-topic",
		Value: sarama.StringEncoder("Hello World"),
	}
	_, _, err = producer.SendMessage(msg)
	if err != nil {
		panic(err)
	}
}

func Example_consumer() {
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, nil)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	consumer = saramatrace.WrapConsumer(consumer)

	partitionConsumer, err := consumer.ConsumePartition("some-topic", 0, sarama.OffsetNewest)
	if err != nil {
		panic(err)
	}
	defer partitionConsumer.Close()

	consumed := 0
	for msg := range partitionConsumer.Messages() {
		// if you want to use the kafka message as a parent span:
		if spanctx, err := tracer.Extract(saramatrace.NewConsumerMessageCarrier(msg)); err == nil {
			// you can create a span using ChildOf(spanctx)
			_ = spanctx
		}

		log.Printf("Consumed message offset %d\n", msg.Offset)
		consumed++
	}
}
