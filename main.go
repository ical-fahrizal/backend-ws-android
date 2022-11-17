package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	conf "api-gateway-android/config"
	"api-gateway-android/router"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {

	router.SetupRoutes()

}

func kafkaCustomer2() {

	//create consumer
	var topics []string
	topics = append(topics, conf.GetTopicRespon())
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	//earliest: automatically reset the offset to the earliest offset
	//latest: automatically reset the offset to the latest offset
	//none: throw exception to the consumer if no previous offset is found for the consumer's group
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               conf.GetBroker(),
		"group.id":                        conf.GetGroup(),
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"auto.offset.reset":               "latest",
	})
	if err != nil {
		log.Printf("Failed to create consumer: %s\n", err)
		os.Exit(1)
	}
	log.Printf("Created Consumer %v\n", c)
	err = c.SubscribeTopics(topics, nil)
	run := true
	for run == true {
		select {
		case sig := <-sigchan:
			log.Printf("Caught signal %v: terminating\n", sig)
			run = false
		case ev := <-c.Events():
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				log.Printf("%% %v\n", e)
				c.Assign(e.Partitions)
			case kafka.RevokedPartitions:
				log.Printf("%% %v\n", e)
				c.Unassign()
			case *kafka.Message:
				//log.Printf("%% Message on %s:\n%s\n",e.TopicPartition, string(e.Value))
				s := string(e.Value)
				log.Printf("respon : %v", s)
				// data := make(map[string]interface{})
				data := &conf.ResponKafka{}
				json.Unmarshal([]byte(s), &data)

				router.SendWs(data)

			case kafka.PartitionEOF:
				log.Printf("%% Reached %v\n", e)
			case kafka.Error:
				log.Printf("Error: %v\n", e)
				run = false

			}
		}
	}

	log.Printf("Closing consumer\n")
	log.Printf("End at %v\n", time.Now())
	c.Close()
	os.Exit(3)
}

//test
