package agent

import (
	"fmt"
	"net"
)

type agent struct{
	conn net.Conn
	stop chan struct{}
	listening bool //true - listening channel exists
}

func NewAgent(addr string) (*agent,  error) {
	conn , err := net.Dial("tcp",addr);
	if err != nil{
		return nil,err
	}
	return &agent{
		conn:   conn,
		stop : make(chan struct{}), // Create the stop channel
	}, nil
}

func (a* agent) listen(){
	for{
		select{
		case <- a.stop:
			return// return when the agent wants to stop listening 
		default:
			fmt.Println("listening...")
		}
	}
}
func (a *agent) BeginListening() {
	a.listening = true
	a.stop = make(chan struct{});
	fmt.Println("started listening to the upcoming messages")
	go a.listen()
}
func (a* agent) Terminate() {
	a.StopListening()
	a.conn.Close()
	fmt.Println("closed connection")
}
func (a* agent) StopListening(){
	if a.listening{
		close(a.stop)
	}
	a.listening = false
	fmt.Println("stopped listening to the upcoming messages")
}
