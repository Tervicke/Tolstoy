package main

import (
	"github.com/Tervicke/Tolstoy/agent"
	"flag"
	"fmt"
	"os"
)
func main(){

	addr := flag.String("addr" ,"localhost:8080" , "The adress to where the Tolstoy server is running")
	topic := flag.String("topic", "mytopic" , "The topic to subscribe to")

	flag.Parse()
	if *addr == "" || *topic == "" {
		fmt.Println(*topic , *addr)
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

	consumer,err:= agent.NewConsumer(*addr , nil)

	fmt.Println("Press Ctrc+C to stop....\nEnter unsubscribe to unsubscribe")

	if err != nil{
		panic(err)
	}
	working := true
	err = consumer.Subscribe(*topic, func(topic , message string){
		fmt.Println("> ",message)
		if message == "unsubscribe" {
			consumer.Unsubscribe(topic)
		}
	})

	if err != nil{
		fmt.Println(err)
		os.Exit(1)
	}
	for working {

	}

}
