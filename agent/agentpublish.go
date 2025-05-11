package agent

import (
	"errors"
	"strings"
	"time"
	"fmt"
)

func (a* agent) Publish(topic string , message string) error{
	packet := makepacket(4,topic,message) //make a packet and then write it
	_, err := a.conn.Write(packet.tobytes())
	if err != nil {
		fmt.Println("Error",err)
	}
	select{
	case ack := <- a.ackchan:
		recieved_topic := strings.Trim(ack.Topic,"\x00")
		recieved_message := strings.Trim(ack.Payload,"\x00")
		if recieved_topic == topic && recieved_message == message{
			return nil
		}else{
			return errors.New(recieved_topic)
		}
	case <-time.After(3 * time.Second):
		return errors.New("Did not recieve ack")
	}
}
