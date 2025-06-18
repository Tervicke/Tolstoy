package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Tervicke/Tolstoy/agent"
)

func main() {

	addr := flag.String("addr", "localhost:8080", "The adress to where the Tolstoy server is running")
	topic := flag.String("topic", "mytopic", "The topic to subscribe to")

	flag.Parse()
	if *addr == "" || *topic == "" {
		fmt.Println(*topic, *addr)
		fmt.Println("Error both the topic and adress are required")
		flag.Usage()
		os.Exit(1)
	}

	/*
		tlsCfg , err := agent.LoadTLSConfig("broker/data/tls/server.crt")

		if err != nil {
			panic(err)
		}
	*/

	consumer, err := agent.NewConsumer(*addr, nil)

	fmt.Println("Press Ctrc+C to stop....\nEnter unsubscribe to unsubscribe")

	if err != nil {
		panic(err)
	}
	working := true
	err = consumer.Subscribe(*topic, func(topic string, payload []byte) {
		message := string(payload)
		fmt.Println("> ", message)
		if message == "unsubscribe" {
			consumer.Unsubscribe(topic)
		}
		if message == "pause" {
			err := consumer.Pause(topic)
			if err != nil {
				fmt.Println("failed to pause" , err)
			}
			fmt.Println("pausing for 5 seconds.....")
			time.Sleep(5 * time.Second)
			err = consumer.Resume(topic , agent.LastKnown) 
			if err != nil {
				fmt.Println("failed to resume")
			}
			fmt.Println("resuming.....")
		}
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for working {

	}

}
func callback(topic , message string){

}
