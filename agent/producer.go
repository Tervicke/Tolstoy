package agent

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
//	"syscall"
	"time"

	pb "github.com/Tervicke/Tolstoy/internal/proto"

	"google.golang.org/protobuf/proto"
)

type producer struct {
	agent
}

//Returns a new producer used to send messages to a topic. 
//tls config should be nil to not use tls settings 
//MaxAttempts for retrying can be overwritten
func NewProducer(addr string, tlsCfg *tls.Config) (*producer, error) {
	var conn net.Conn
	var err error = nil
	if tlsCfg != nil {
		conn, err = tls.Dial("tcp", addr, tlsCfg)
	} else {
		//tls config
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
	p := &producer{}
	p.agent = agent{
		conn:        conn,
		stop:        make(chan struct{}), // Create the stop channel
		ackchannels: make(map[string]chan *pb.Packet),
		MaxAttempts: 3,
		serverAddr:  addr,
		tlsCfg:      tlsCfg,
		listener: p,
	}
	go p.listen()
	return p, nil
}

//Used to publish a message in a particular topic
func (p *producer) Publish(topic string, payload []byte) error {
	Id := generateUniqueId(p.ackchannels)
	p.ackchannels[Id] = make(chan *pb.Packet)
	defer delete(p.ackchannels, Id)

	pubPacket := &pb.Packet{
		Type:      pb.Type_PUBLISH,
		Topic:     topic,
		Payload:   payload,
		RequestId: Id,
	}
	err := p.safeWritePacket(pubPacket) 
	if err != nil{
		return err
	}
	for i := 1; i <= p.MaxAttempts; i++ {
		select {
		case ack := <-p.ackchannels[Id]:
			if ack.Type == pb.Type_ERROR {
				return errors.New(ack.Error.Text)
			} else {
				return nil
			}
		case <-time.After(500 * time.Millisecond):
			//log.Println("failed to recieve the ack, retrying..", i) //test
		}
		time.Sleep(time.Duration((i * 100)) * time.Millisecond) //linear backoff
	}	
	return fmt.Errorf("Error recieving packet ack tried %d times", p.MaxAttempts)
}
