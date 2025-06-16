package agent

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
	"time"

	pb "github.com/Tervicke/Tolstoy/internal/proto"

	"google.golang.org/protobuf/proto"
)

type producer struct {
	conn        net.Conn
	stop        chan struct{}
	listening   bool                       //true - listening channel exists
	callbacks   map[string]OnMessage       //map for the callbacks of various topics
	ackchannels map[string]chan *pb.Packet //channel for handling ack
	MaxAttempts int
	serverAddr  string      //addr of the connected server
	tlsCfg      *tls.Config //tls config if any nil if none
}

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
	p := &producer{
		conn:        conn,
		stop:        make(chan struct{}), // Create the stop channel
		ackchannels: make(map[string]chan *pb.Packet),
		MaxAttempts: 3,
		serverAddr:  addr,
		tlsCfg:      tlsCfg,
	}
	go p.listen()
	return p, nil
}

func (p *producer) listen() {
	for {
		select {
		case <-p.stop:
			return
		default:
			//read the size first and then read the buf and deserialize it
			sizeBuf := make([]byte, 4)
			io.ReadFull(p.conn, sizeBuf)
			size := binary.BigEndian.Uint32(sizeBuf)
			msgBuf := make([]byte, size)
			io.ReadFull(p.conn, msgBuf)
			//serialize
			packet := &pb.Packet{}
			proto.Unmarshal(msgBuf, packet)
			switch packet.Type {
			default:
				p.ackchannels[packet.RequestId] <- packet
			}
		}
	}
}

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

func (p *producer) Terminate() error {
	//send the Disconnection Packet
	disConPacket := &pb.Packet{
		Type: pb.Type_DIS_CONN_REQUEST,
	}
	err := p.safeWritePacket(disConPacket) 
	if err != nil {
		return err
	}

	p.StopListening()
	p.conn.Close()
	return nil
}

func (p *producer) StopListening() {
	if p.listening {
		close(p.stop)
	}
	p.listening = false
}

// refers to the broken pipe returns true if broken pipe fixed after certain attemps else returns false
func (p *producer) brokenPipe() bool {
	for i := 1; i <= p.MaxAttempts; i++ {
		var err error = nil
		var conn net.Conn
		if p.tlsCfg != nil {
			conn, err = tls.Dial("tcp", p.serverAddr, p.tlsCfg)
		} else {
			//tls config
			conn, err = net.Dial("tcp", p.serverAddr)
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
		p.conn = conn
		go p.listen()
		return true
	}
	return false
}
func (p *producer)safeWritePacket(packet *pb.Packet) (error){
	for i := 1; i <= p.MaxAttempts; i++ {
		err := writePacket(p.conn, packet)
	
		if err != nil {
			//reconnect and try to send the message then
			if errors.Is(err, syscall.EPIPE) {
				fixed := p.brokenPipe()
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
