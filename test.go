package main

import (
	"Tolstoy/agent"
	"fmt"
	"time"
)

func main(){
	addr := "localhost:8080"
	agent,err := agent.NewAgent(addr)
	if err != nil{
		panic(err)
	}
	agent.BeginListening()
	time.Sleep(1 * time.Second)
	agent.StopListening()
	time.Sleep(2 * time.Second)
	fmt.Println("doing it all over again")
	time.Sleep(1 * time.Second)
	agent.BeginListening()
	time.Sleep(1 * time.Second)
	agent.Terminate()
}
