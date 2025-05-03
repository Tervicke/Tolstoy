package main

import (
	"Tolstoy/agent"
	"fmt"
	"time"
)


func makeandpublish(){
	agent,err := agent.NewAgent("localhost:8080")
	if err != nil {
		panic(err)
	}
	now := time.Now().UnixNano() // nanosecond-precision timestamp
	msg := fmt.Sprintf("%d", now) // Just the timestamp as message
	err = agent.Publish("mytopic", msg)
	if err != nil {
		fmt.Println("Publish error:", err)
	}
	time.Sleep(1 * time.Second)
}

func main(){
	for {
		go makeandpublish()
		time.Sleep(1 * time.Second)
	}
}
