package agent

import (
	"sync"

	pb "github.com/Tervicke/Tolstoy/internal/proto"

	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"google.golang.org/protobuf/proto"
)

type consumer struct {
	conn        net.Conn
	stop        chan struct{}
	listening   bool                         //true - listening channel exists
	callbacks   map[string]OnMessage         //map for the callbacks of various topics
	ackchannels map[string]chan *pb.Packet   //keep a list of ack channels associated with their unique Id
	offsets     map[string]int64             //map to store string and offsets
	workqueue   map[string](chan *pb.Packet) //used to store the last recieved packets and run them concurrently and orderly
	mu sync.Mutex
}

func NewConsumer(addr string, tlsCfg *tls.Config) (*consumer, error) {
	var conn net.Conn
	var err error
	if tlsCfg != nil {
		conn, err = tls.Dial("tcp", addr, tlsCfg)
	} else {
		conn, err = net.Dial("tcp", addr)
	}

	if err != nil {
		return nil, err
	}
	connPacket := &pb.Packet{
		Type: pb.Type_CONN_REQUEST,
	}
	writePacket(conn, connPacket)

	//read the conn request ack
	sizeBuf := make([]byte, 4)
	_, err = io.ReadFull(conn, sizeBuf)
	if err != nil {
		fmt.Println("packet size error")
		return nil, errors.New("server did not send correct packet size response")
	}
	size := binary.BigEndian.Uint32(sizeBuf)
	msgBuf := make([]byte, size)
	_, err = io.ReadFull(conn, msgBuf)
	if err != nil {
		fmt.Println("packet error")
		return nil, errors.New("server did not send correct response")
	}
	packet := &pb.Packet{}
	err = proto.Unmarshal(msgBuf, packet)
	if err != nil {
		return nil, errors.New("Failed to make consumer")
	}
	if packet.Type != pb.Type_ACK_CONN_REQUEST {
		return nil, errors.New("Failed to connect to server")
	}
	c := &consumer{
		conn:        conn,
		stop:        make(chan struct{}), // Create the stop channel
		ackchannels: make(map[string]chan *pb.Packet),
		offsets:     make(map[string]int64),
		workqueue:   make(map[string](chan *pb.Packet)),
	}
	go c.listen()
	return c, nil
}

func (c *consumer) listen() {
	for {
		select {
		case <-c.stop:
			return
		default:
			//read the size first and then read the buf and deserialize it
			sizeBuf := make([]byte, 4)
			io.ReadFull(c.conn, sizeBuf)
			size := binary.BigEndian.Uint32(sizeBuf)
			msgBuf := make([]byte, size)
			io.ReadFull(c.conn, msgBuf)
			//serialize
			packet := &pb.Packet{}
			proto.Unmarshal(msgBuf, packet)
			switch packet.Type {
			case pb.Type_DELIVER:
				//store the offset 
				c.offsets[packet.Topic] = packet.Offset
				
				//get the callback
				//callbackfunction := c.callbacks[packet.Topic]

				//calling the function
				//go callbackfunction(packet.Topic, packet.Payload)
				c.workqueue[packet.Topic]<-packet  
			default:
				c.ackchannels[packet.RequestId] <- packet
			}
		}
	}
}

func (c *consumer) processWorkQueue(topic string) {
	for packet := range c.workqueue[topic] {

		function := c.callbacks[packet.Topic]
		function(topic , packet.Payload)

	}
}

func (c *consumer) Subscribe(topic string, callback OnMessage) error {
	Id := generateUniqueId(c.ackchannels)
	c.ackchannels[Id] = make(chan *pb.Packet)
	defer delete(c.ackchannels, Id)

	subpacket := &pb.Packet{
		Type:      pb.Type_SUBSCRIBE,
		Topic:     topic,
		RequestId: Id,
	}

	err := writePacket(c.conn, subpacket)
	if err != nil {
		return errors.New("Failed to send subscribe packet")
	}
	select {
	case ack := <-c.ackchannels[Id]:

		if ack.Type == pb.Type_ERROR {
			return errors.New(pb.Type_ERROR.String())
		} else {
			if c.callbacks == nil {
				c.callbacks = make(map[string]OnMessage)
			}
			c.callbacks[topic] = callback
			if _ , exists := c.workqueue[topic]; !exists{
				c.workqueue[topic] = make(chan *pb.Packet,100) 
				go c.processWorkQueue(topic)
			}
			return nil
		}

	case <-time.After(3 * time.Second):
		return errors.New("Did not recieve ack")
	}
}

func (c *consumer) Unsubscribe(topic string) error {
	Id := generateUniqueId(c.ackchannels)
	c.ackchannels[Id] = make(chan *pb.Packet)
	defer delete(c.ackchannels, Id)

	unSubPacket := &pb.Packet{
		Type:      pb.Type_UNSUBSCRIBE,
		Topic:     topic,
		RequestId: Id,
	}
	err := writePacket(c.conn, unSubPacket)
	if err != nil {
		return err
	}
	select {
	case ack := <-c.ackchannels[Id]:
		if ack.Type != pb.Type_ERROR {
			return errors.New(ack.Error.String())
		} else {
			return nil
		}
	case <-time.After(3 * time.Second):
		return errors.New("did not recieve ack")
	}
}
func (c *consumer) Terminate() error {
	//send the Disconnection Packet
	disConPacket := &pb.Packet{
		Type: pb.Type_DIS_CONN_REQUEST,
	}
	err := writePacket(c.conn, disConPacket)

	if err != nil {
		return err
	}

	c.StopListening()
	c.conn.Close()
	return nil
}

func (c *consumer) StopListening() {
	if c.listening {
		close(c.stop)
	}
	c.listening = false
}

func (c *consumer) Pause(topic string) error {
	Id := generateUniqueId(c.ackchannels)
	c.ackchannels[Id] = make(chan *pb.Packet)
	defer delete(c.ackchannels, Id)

	pausePacket := &pb.Packet{
		Type:      pb.Type_PAUSE,
		Topic:     topic,
		RequestId: Id,
	}
	err := writePacket(c.conn, pausePacket)
	if err != nil {
		return err
	}
	select {
	case ack := <-c.ackchannels[Id]:
		if ack.Type == pb.Type_ERROR {
			return errors.New(ack.GetError().Text)
		} else {
			return nil
		}
	case <-time.After(3 * time.Second):
		return errors.New("Did not recieve ack")
	}
}

type ResumeMode int

const (
	LastKnown ResumeMode = iota
	Latest
)

func (c *consumer) Resume(topic string, mode ResumeMode) error {
	Id := generateUniqueId(c.ackchannels)
	c.ackchannels[Id] = make(chan *pb.Packet)
	defer delete(c.ackchannels, Id)
	var offset int64
	switch mode {
	case Latest:
		offset = -1
	case LastKnown:
		offset = c.offsets[topic]
	}
	resumePacket := &pb.Packet{
		Type:      pb.Type_RESUME,
		Topic:     topic,
		RequestId: Id,
		Offset:    offset,
	}
	fmt.Println("resuming from ", offset)

	err := writePacket(c.conn, resumePacket)

	if err != nil {
		return err
	}
	select {
	case ack := <-c.ackchannels[Id]:
		if ack.Type == pb.Type_ERROR {
			return errors.New(ack.GetError().Text)
		} else {
			return nil
		}
	case <-time.After(3 * time.Second):
		return errors.New("Did not recieve ack")
	}
}
