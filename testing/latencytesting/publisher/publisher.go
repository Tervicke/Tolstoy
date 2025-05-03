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
	for {
		now := time.Now().UnixNano() 		
		msg := fmt.Sprintf("%d", now) 
		err = agent.Publish("mytopic", msg)
		if err != nil {
			fmt.Println("Publish error:", err)
		}
		//time.Sleep(time.Nanosecond)
	}
}

func main(){
	makeandpublish()
}
