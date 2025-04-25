package main

import (
	"Tolstoy/agent"
	"fmt"
	"flag"
	"os"
)

func main(){
	addr := flag.String("addr","","The adress to connect to")
	topic := flag.String("topic","","The topic to send messages to")
	flag.Parse()
	if *addr == "" || *topic == "" {
		fmt.Println("Error: both addr and topic are required.")
		flag.Usage()
		os.Exit(1)   
	}
	agent,err := agent.NewAgent(*addr)
	if err != nil{
		panic(err)
	}
	var message string;
	for {
		fmt.Scan(&message)
		if message == "exit" {
			os.Exit(0)
		}
		agent.Publish(*topic, message)
	}
}
