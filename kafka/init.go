package kafka

import (
	"fmt"
	"log"

	"github.com/Shopify/sarama"
	"gopkg.in/urfave/cli.v2"
)

var (
	kAsyncProducer sarama.AsyncProducer
	kClient        sarama.Client
	ChatTopic      string
)

func initKafka(addrs []string, chatTopic string) {
	ChatTopic = chatTopic
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	producer, err := sarama.NewAsyncProducer(addrs, config)
	if err != nil {
		log.Fatalln(err)
	}

	kAsyncProducer = producer
	cli, err := sarama.NewClient(addrs, nil)
	if err != nil {
		log.Fatalln(err)
	}
	kClient = cli
}

func Init(c *cli.Context) {
	addrs := c.StringSlice("kafka-brokers")
	topic := c.String("chat-topic")
	initKafka(addrs, topic)
}

func InitTest(addrs []string, chatTopic string) {
	initKafka(addrs, chatTopic)
}

func NewConsumer() (sarama.Consumer, error) {
	return sarama.NewConsumerFromClient(kClient)
}

func SendChat(ep uint64, msg []byte) {
	pm := &sarama.ProducerMessage{
		Topic: ChatTopic,
		Key:   sarama.StringEncoder(fmt.Sprint(ep)),
		Value: sarama.ByteEncoder(msg),
	}
	kAsyncProducer.Input() <- pm
	select {
	case e := <-kAsyncProducer.Errors():
		log.Println("error", ChatTopic, ep, e)
	case <-kAsyncProducer.Successes():
		// log.Println("succe", s)
		// default:
	}
}
