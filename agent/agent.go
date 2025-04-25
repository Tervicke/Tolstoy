package agent

import (
	"net"
)

type agent struct{
	conn net.Conn
	stop chan struct{}
	listening bool //true - listening channel exists
	ackchan chan packet
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
			if(packet.Type == 10){ //ack code - 10 = publish ack 
				a.ackchan <- packet
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
