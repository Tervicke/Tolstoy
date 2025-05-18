package main

import (
	"Tolstoy/agent"
	"fmt"
	"math"
	"sort"
	"strconv"
	"text/tabwriter"
	"time"
	"os"
)

func addPublisher(total_messages int){
	agent,err := agent.NewProducer("localhost:8080")
	if err != nil {
		panic(err)
	}
	defer agent.Terminate()
	var current_message = 1;
	for current_message <= total_messages{
		now := time.Now().UnixNano() 		
		msg := fmt.Sprintf("%d", now) 
		err = agent.Publish("mytopic", msg)
		if err != nil {
			fmt.Println("Publish error:", err)
			panic(err)
		}
		current_message++
	}
}

func addSubscriber(total_messages int , done chan struct{}){

	sub, err := agent.NewConsumer("localhost:8080")
	
	if err != nil {
		panic(err)
	}

	var sum float64
	var latencies []float64;
	messagesdone := make(chan struct{})

	var count = 1;

	err = sub.Subscribe("mytopic", func(topic , message string){
		//get the current time stamp
		sentTime, err := strconv.ParseInt(message, 10, 64)
		if err != nil{
			fmt.Println("recieved" , message)
			fmt.Println("-------------------------------------");
			return
		} 

		now := time.Now().UnixNano()
		latencyMs := float64(now - sentTime) / 1_000_000.0
		
		sum += latencyMs

		latencies = append(latencies, latencyMs)

		count++
		if count == total_messages {
			close(messagesdone)
		}

	})

	<-messagesdone

	//calculate standard devitation	
	var variance_sum float64 = 0;

	var average_latency float64 = sum / float64(total_messages) 

	for _,latency := range latencies {
		diff_square := math.Pow( (latency - average_latency) , 2.0)
		variance_sum += diff_square / float64(total_messages - 1); 
	}

	var variance float64 = math.Sqrt(variance_sum)


	//calculate the p95 and p99
	sort.Float64s(latencies)
	p95_index := int(0.95 * float64(total_messages))
	p99_index := int(0.99 * float64(total_messages))

	printStats(average_latency,variance,latencies[p95_index] , latencies[p99_index])
	sub.Terminate()

	close(done)
}

func TestLatency(total_messages int){
	done := make(chan struct{})
	go addSubscriber(total_messages,done)
	go addPublisher(total_messages)
	fmt.Println("Baseline Test Result (10000 messages transfered)")
	<-done
}

func printStats(average_latency, variance, p95, p99 float64) {
    writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
    fmt.Fprintln(writer, "Metric\tValue")
    fmt.Fprintf(writer, "Average\t%.4f ms\n", average_latency)
    fmt.Fprintf(writer, "Jitter\t%.4f ms\n", variance)
    fmt.Fprintf(writer, "P95\t%.4f ms\n", p95)
    fmt.Fprintf(writer, "P99\t%.4f ms\n", p99)
    writer.Flush()  // Don't forget to flush to output the results!
}
