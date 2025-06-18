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
	"syscall"

	"google.golang.org/protobuf/proto"
)

// consumer manages the client-side message consumption lifecycle in the pub/sub system.
// It maintains network connection state, topic subscriptions, message acknowledgment,
// and ordered processing through per-topic work queues.
//MaxAttempts for retrying can be overwritten
type consumer struct {
	conn        net.Conn
	stop        chan struct{}
	listening   bool                         //true - listening channel exists
	callbacks   map[string]OnMessage         //map for the callbacks of various topics
	ackchannels map[string]chan *pb.Packet   //keep a list of ack channels associated with their unique Id
	offsets     map[string]int64             //map to store string and offsets
	workqueue   map[string](chan *pb.Packet) //used to store the last recieved packets and run them concurrently and orderly
	mu sync.Mutex
	MaxAttempts int
	serverAddr  string      //addr of the connected server
	tlsCfg      *tls.Config //tls config if any nil if none
}

//returns a instance of agent.consumer based on the addr and the tls config if provided
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
		MaxAttempts: 3,
		serverAddr:  addr,
		tlsCfg:      tlsCfg,
	}
	go c.listen()
	return c, nil
}

//The internal function listen that runs in a go routine and listens to the incoming packets also routes them 
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

//Internal function which processes the callbacks in a ordered manner
func (c *consumer) processWorkQueue(topic string) {
	for packet := range c.workqueue[topic] {

		function := c.callbacks[packet.Topic]
		function(topic , packet.Payload)

	}
}

//Used to subscribe to a a topic , the callback function is runs whenever a message is recieved
func (c *consumer) Subscribe(topic string, callback OnMessage) error {
	Id := generateUniqueId(c.ackchannels)
	c.ackchannels[Id] = make(chan *pb.Packet)
	defer delete(c.ackchannels, Id)

	subpacket := &pb.Packet{
		Type:      pb.Type_SUBSCRIBE,
		Topic:     topic,
		RequestId: Id,
	}

	err := c.safeWritePacket(subpacket)
	if err != nil {
		return err
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

//Used to unsubscribe from a particular topic , the broker doesnt return an error if u try to unsubscribe from a non subscribed topic
func (c *consumer) Unsubscribe(topic string) error {
	Id := generateUniqueId(c.ackchannels)
	c.ackchannels[Id] = make(chan *pb.Packet)
	defer delete(c.ackchannels, Id)

	unSubPacket := &pb.Packet{
		Type:      pb.Type_UNSUBSCRIBE,
		Topic:     topic,
		RequestId: Id,
	}
	err := c.safeWritePacket(unSubPacket)
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

//Terminates the consumer instance
func (c *consumer) Terminate() error {
	//send the Disconnection Packet
	disConPacket := &pb.Packet{
		Type: pb.Type_DIS_CONN_REQUEST,
	}
	err := c.safeWritePacket(disConPacket)

	if err != nil {
		return err
	}

	c.stoplistening()
	c.conn.Close()
	return nil
}

//Internal used to stop listening to the server and terminate the listen go routine
func (c *consumer) stoplistening() {
	if c.listening {
		close(c.stop)
	}
	c.listening = false
}

//Pause stops the Incoming messages from a channel.
//The topic is still subscribed.
func (c *consumer) Pause(topic string) error {
	Id := generateUniqueId(c.ackchannels)
	c.ackchannels[Id] = make(chan *pb.Packet)
	defer delete(c.ackchannels, Id)

	pausePacket := &pb.Packet{
		Type:      pb.Type_PAUSE,
		Topic:     topic,
		RequestId: Id,
	}
	err := c.safeWritePacket(pausePacket)
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

// Resume starts consuming messages from the given topic based on the specified mode.
// In Latest mode, the consumer ignores older messages and starts from the most recent.
// In LastKnown mode, the consumer resumes from the last acknowledged offset, if available.
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

	err := c.safeWritePacket(resumePacket)

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

//Internal safe write packet used to check and fix error when writing packet
func (c *consumer)safeWritePacket(packet *pb.Packet) (error){
	for i := 1; i <= c.MaxAttempts; i++ {
		err := writePacket(c.conn, packet)
	
		if err != nil {
			//reconnect and try to send the message then
			if errors.Is(err, syscall.EPIPE) {
				fixed := c.brokenPipe()
				if !fixed {
					return errors.New("failed to send the packet,broken pipe")
				}else{
					continue
				}
			}
			return err
		}else{
			break
		}
	}
	return nil
}

func (c *consumer) brokenPipe() bool {
	for i := 1; i <= c.MaxAttempts; i++ {
		var err error = nil
		var conn net.Conn
		if c.tlsCfg != nil {
			conn, err = tls.Dial("tcp", c.serverAddr, c.tlsCfg)
		} else {
			//tls config
			conn, err = net.Dial("tcp", c.serverAddr)
		}
		if err != nil {
			//linearly wait for the timeout
			time.Sleep(500 * time.Millisecond)
			continue
		}
		connPacket := &pb.Packet{
			Type: pb.Type_CONN_REQUEST,
		}
		err = writePacket(conn, connPacket)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		//read the conn request ack
		sizeBuf := make([]byte, 4)
		_, err = io.ReadFull(conn, sizeBuf)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		size := binary.BigEndian.Uint32(sizeBuf)
		msgBuf := make([]byte, size)
		_, err = io.ReadFull(conn, msgBuf)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		packet := &pb.Packet{}
		err = proto.Unmarshal(msgBuf, packet)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if packet.Type != pb.Type_ACK_CONN_REQUEST {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		c.conn = conn
		go c.listen()
		return true
	}
	return false
}
