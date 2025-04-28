package agent

import (
	"net"
	"strings"
)

type agent struct{
	conn net.Conn
	stop chan struct{}
	listening bool //true - listening channel exists
	ackchan chan Packet //publish ack channel
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
		ackchan:  make(chan Packet),
	}
	go a.listen()
	return a,nil
}

type onMessage func(topic string, message string)


func (a *agent) Unsubscribe(topic string){
	unsubpacket := makepacket(6,topic,"")
	a.conn.Write(unsubpacket.tobytes())
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

