package main
import (
	"Tolstoy/agent"
	"fmt"
	"time"
	"strconv"
)

func main() {
	// Connect the subscriber agent
	sub, err := agent.NewAgent("localhost:8080")
	if err != nil {
		panic(err)
	}
	var sum float64
	var count = 0;
	var average float64 = 0;
	err = sub.Subscribe("mytopic", func(topic , message string){
		//get the current time stamp
		sentTime, err := strconv.ParseInt(message, 10, 64)
		if err != nil{
			fmt.Println("recieved" , message)
			return
		} 

		now := time.Now().UnixNano()
		latencyMs := float64(now - sentTime) / 1_000_000.0
		
		sum += latencyMs
		count++


		average = sum / float64(count) 
		fmt.Printf("Average Latency over %d : %.5f ms\n",count ,average)
	})

	defer sub.Terminate()

	select {}
}

