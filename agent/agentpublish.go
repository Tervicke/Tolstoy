package agent

import (
	"errors"
	"strings"
	"time"
	"fmt"
)

func (a* agent) Publish(topic string , message string) error{
	buf := make([]byte,2049)
	buf[0] = 4; 
	copy(buf[1:1025] , []byte(topic))
	copy(buf[1025:2049] , []byte(message))
	go func() {
        _, err := a.conn.Write(buf[:])
        if err != nil {
						fmt.Println("Error occured recieved wrong ack")
						return
        }
    }()
	//	a.conn.Write(buf[:])
	select{
	case ack := <- a.ackchan:
		recieved_topic := strings.Trim(ack.Topic,"\x00")
		recieved_message := strings.Trim(ack.Payload,"\x00")
		if recieved_topic == topic && recieved_message == message{
			return nil
		}else{
			return errors.New("Error occured recieved wrong ack")
		}
	case <-time.After(3 * time.Second):
		return errors.New("Did not recieve ack")
	}
}
