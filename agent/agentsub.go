package agent
import (
	"strings"
	"errors"
	"time"
)
func (a *agent) Subscribe(topic string , callback onMessage) (error){
	subpacket := makepacket(5,topic,"")  
	a.conn.Write(subpacket.tobytes())
	select{
	case ack := <- a.ackchan:
		recieved_topic := strings.Trim(ack.Topic,"\x00")
		if recieved_topic == topic{
			if a.callbacks == nil{
				a.callbacks = make(map[string]onMessage)
			}
			a.callbacks[topic] = callback
			return nil
		}else{
			return errors.New(string(ack.Topic))
		}
	case <-time.After(3 * time.Second):
		return errors.New("Did not recieve ack")
	}
}
