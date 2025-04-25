package agent

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

func (a* agent) Publish(topic string , message string) error{
	buf := make([]byte,2049)
	buf[0] = 4; 
	copy(buf[1:2025] , []byte(topic))
	copy(buf[1025:2049] , []byte(message))
	a.conn.Write(buf[:])
	fmt.Println("message sent from agent")
	select{
	case ack := <- a.ackchan:
		recieved_topic := strings.Trim(ack.Topic,"\x00")
		recieved_message := strings.Trim(ack.Payload,"\x00")
		fmt.Println("------------------------------")
		fmt.Println(topic , len(topic) , )
		fmt.Println(ack.Topic , len(recieved_topic) )
		fmt.Println("------------------------------")
		if recieved_topic == topic && recieved_message == message{
			fmt.Println("recived the ack!!!!!!!!!!")
		}else{
			fmt.Println("hello ????")
		}
	case <-time.After(3 * time.Second):
		return errors.New("Did not recieve ack")
	}
	return nil
}
