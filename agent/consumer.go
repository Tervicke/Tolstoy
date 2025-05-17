package agent

import (
	pb "Tolstoy/proto"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"google.golang.org/protobuf/proto"
)
type consumer struct{
	conn net.Conn
	stop chan struct{}
	listening bool //true - listening channel exists
	ackchan chan *pb.Packet //publish ack channel
	callbacks map[string]OnMessage //map for the callbacks of various topics
}


func NewConsumer(addr string) (*consumer , error) {
	conn , err := net.Dial("tcp",addr);
	if err != nil{
		return nil,err
	}
	connPacket := &pb.Packet{
		Type: pb.Type_CONN_REQUEST,
	}
	writePacket(conn , connPacket)

	//read the conn request ack
	sizeBuf := make([]byte,4)
	_ , err = io.ReadFull(conn , sizeBuf)
	if err != nil {
		fmt.Println("packet size error")
		return nil,errors.New("server did not send correct packet size response")
	}
	size := binary.BigEndian.Uint32(sizeBuf)
	msgBuf := make([]byte,size)
	_, err = io.ReadFull(conn , msgBuf)
	if err != nil {
		fmt.Println("packet error")
		return nil,errors.New("server did not send correct response")
	}
	packet := &pb.Packet{}
	err = proto.Unmarshal(msgBuf , packet)
	if err != nil {
		return nil,errors.New("Failed to make consumer")
	}
	if packet.Type != pb.Type_ACK_CONN_REQUEST {
		return nil,errors.New("Failed to connect to server")
	}
	c := &consumer{
		conn : conn,
		stop : make(chan struct{}), // Create the stop channel
		ackchan:  make(chan *pb.Packet),
	}
	go c.listen()
	return c,nil
}

func (c *consumer) listen(){
	for {
		select {
			case <- c.stop:
				return
			default:
				//read the size first and then read the buf and deserialize it
				sizeBuf := make([]byte,4)
				io.ReadFull(c.conn , sizeBuf)
				size := binary.BigEndian.Uint32(sizeBuf)
				msgBuf := make([]byte,size)
				io.ReadFull(c.conn , msgBuf)
				//serialize
				packet := &pb.Packet{}
				proto.Unmarshal(msgBuf , packet)
				switch packet.Type{
					case pb.Type_ACK_SUBSCRIBE:
						c.ackchan <- packet
					case pb.Type_DELIVER:
						//callback the function on recieved topic
						callbackfunction := c.callbacks[packet.Topic]
						//calling the function
						callbackfunction(packet.Topic , packet.Payload)
				}
		}
	}
}

func (c *consumer) Subscribe(topic string , callback OnMessage) (error){
	subpacket := &pb.Packet{
		Type: pb.Type_SUBSCRIBE,
		Topic: topic,
	}
	err := writePacket(c.conn , subpacket)
	if err != nil {
		return errors.New("Failed to send subscribe packet")
	}
	select{

		case ack := <-c.ackchan:
			if ack.Type == pb.Type_ACK_SUBSCRIBE && ack.Topic == topic{

				//save the callback
				if c.callbacks == nil{
					c.callbacks = make(map[string]OnMessage)
				}
				c.callbacks[topic] = callback

				return nil
			}
			return errors.New("Recieved wrong ack")

		case <-time.After(3 * time.Second):
			return errors.New("Did not recieve ack")

	}
}

func (c *consumer) Unsubscribe(topic string) (error){
	unSubPacket :=&pb.Packet{
		Type: pb.Type_UNSUBSCRIBE,
		Topic: topic,
	} 
	err := writePacket(c.conn , unSubPacket)
	if err != nil {
		return err
	}
	return nil
}
func (c *consumer) Terminate() (error){
	//send the Disconnection Packet
	disConPacket := &pb.Packet{
		Type : pb.Type_DIS_CONN_REQUEST,
	}
	err := writePacket(c.conn , disConPacket)

	if err != nil {
		return err
	}

	c.StopListening()
	c.conn.Close()
	return nil
}

func (c *consumer) StopListening(){
	if c.listening{
		close(c.stop)
	}
	c.listening = false
}
