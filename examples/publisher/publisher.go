package main

import (
	"Tolstoy/agent"
	"bufio"
	"fmt"
	"flag"
	"os"
)

func main() {
	// Initialize the agent
	addr := flag.String("addr" ,"localhost:8080" , "The adress to where the Tolstoy server is running")
	topic := flag.String("topic", "mytopic" , "The topic to send message to")
	flag.Parse()
	if *addr == "" || *topic == "" {
		fmt.Println(*topic , *addr)
		fmt.Println("Error both the topic and adress are required")
		flag.Usage()
		os.Exit(1)
	}
	producer, err := agent.NewProducer(*addr)
	fmt.Println("type exit to exit")
	if err != nil {
		panic(err)
	}
	//defer agent.Terminate()
	// Create a scanner for reading input
	scanner := bufio.NewScanner(os.Stdin)

	// Reading the input line by line
	for {
		fmt.Print("> ")
		if scanner.Scan() {
			message := scanner.Text()
			if message == "exit" {
				os.Exit(0)
			}
			err := producer.Publish(*topic, message)
			if err != nil {
				fmt.Println("Error occurred:", err)
				break
			} 
		} else {
			if err := scanner.Err(); err != nil {
				fmt.Println("Error reading input:", err)
				break
			}
			break
		}
	}
}

