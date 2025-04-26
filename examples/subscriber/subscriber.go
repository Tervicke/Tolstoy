package main

import (
	"Tolstoy/agent"
	"flag"
	"fmt"
	"os"
)
func main(){

	addr := flag.String("addr" ,"" , "The adress to where the Tolstoy server is running")
	topic := flag.String("topic", "" , "The topic to subscribe to")

	flag.Parse()
	if *addr == "" || *topic == "" {
		fmt.Println(*topic , *addr)
		fmt.Println("Error both the topic and adress are required")
		flag.Usage()
		os.Exit(1)
	}

	agent,err:= agent.NewAgent(*addr)
	fmt.Println("Press Ctrc+C to stop....")
	if err != nil{
		panic(err)
	}
	defer agent.Terminate()
	err = agent.Subscribe(*topic, func(topic , message string){
		fmt.Println("> ",message)
	})
	if err != nil{
		fmt.Println(err)
		os.Exit(1)
	}
	for {
	}
}
