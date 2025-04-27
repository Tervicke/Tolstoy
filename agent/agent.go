package agent

import (
	"net"
	"strings"
	"errors"
	"time"
)

type agent struct{
	conn net.Conn
	stop chan struct{}
	listening bool //true - listening channel exists
	ackchan chan packet //publish ack channel
	callbacks map[string]onMessage //map for the callbacks of various topics
}

func NewAgent(addr string) (*agent,  error) {
	conn , err := net.Dial("tcp",addr);
	if err != nil{
		return nil,err
	}
	a := &agent{
		conn:   conn,
		stop : make(chan struct{}), // Create the stop channel
		ackchan:  make(chan packet),
	}
	go a.listen()
	return a,nil
}

type onMessage func(topic string, message string)

func (a *agent) Subscribe(topic string , callback onMessage) (error){
	buf := make([]byte,2049)
	buf[0] = 5; 
	copy(buf[1:1025] , []byte(topic))
	a.conn.Write(buf[:])

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

func (a *agent) Unsubscribe(topic string){
	buf := make([]byte,2049) 
	buf[0] = 6; 
	copy(buf[1:1025] , []byte(topic))
	a.conn.Write(buf[:])
}


func (a* agent) listen(){
	for{
		select {
			case <- a.stop:
				return
			default:

			buf := make([]byte,2049)
			totalread := 0;
			for totalread < 2049 {
				n,_ := a.conn.Read(buf[totalread:])
				totalread += n
			}
			packet := newpacket([2049]byte(buf))
			switch packet.Type{
				case 10,11,2:
					a.ackchan <- packet
				case 12:
					return
				case 3:
					recieved_topic := strings.Trim(packet.Topic,"\x00")
					if callback , exists := a.callbacks[recieved_topic] ; exists{
						callback(strings.Trim(packet.Topic,"\x00"),strings.Trim(packet.Payload,"\x00"))
					}
			}
		} 
	}
}

func (a* agent) Terminate() {
	a.StopListening()
	a.conn.Close()
}
func (a* agent) StopListening(){
	if a.listening{
		close(a.stop)
	}
	a.listening = false
}
