package agent

import (
	pb "github.com/Tervicke/Tolstoy/internal/proto"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
	"crypto/tls"

	"google.golang.org/protobuf/proto"
)

type producer struct{
	conn net.Conn
	stop chan struct{}
	listening bool //true - listening channel exists
	callbacks map[string]OnMessage //map for the callbacks of various topics
	ackchannels map[string]chan *pb.Packet //channel for handling ack
}

func NewProducer(addr string , tlsCfg *tls.Config) (*producer , error) {
	var conn net.Conn
	var err error = nil
	if tlsCfg != nil {
		conn , err = tls.Dial("tcp",addr , tlsCfg);
	}else{
		//tls config
		conn , err = net.Dial("tcp",addr);
	}
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
	p := &producer{
		conn : conn,
		stop : make(chan struct{}), // Create the stop channel
		ackchannels:  make(map[string]chan *pb.Packet),
	}
	go p.listen()
	return p,nil
}

func (p *producer) listen(){
	for {
		select {
			case <- p.stop:
				return
			default:
				//read the size first and then read the buf and deserialize it
				sizeBuf := make([]byte,4)
				io.ReadFull(p.conn , sizeBuf)
				size := binary.BigEndian.Uint32(sizeBuf)
				msgBuf := make([]byte,size)
				io.ReadFull(p.conn , msgBuf)
				//serialize
				packet := &pb.Packet{}
				proto.Unmarshal(msgBuf , packet)
				switch packet.Type{
						default :
						p.ackchannels[packet.RequestId]<-packet
				}
		}
	}
}

func (p *producer) Publish(topic , payload string) (error){

	Id := generateUniqueId(p.ackchannels)
	p.ackchannels[Id] = make(chan *pb.Packet)
	defer delete(p.ackchannels , Id)

	pubPacket := &pb.Packet{
		Type:pb.Type_PUBLISH,
		Topic : topic,
		Payload: payload,
		RequestId: Id,
	}
	err := writePacket(p.conn , pubPacket)
	if err != nil {
		return err
	}
	select {
	case ack := <-p.ackchannels[Id]:
		if ack.Type == pb.Type_ERROR{
			return errors.New(ack.Error.Text)
		}else{
			return nil
		}
	case <- time.After(3 * time.Second):
		return errors.New("Did not recieve ack")
	}
}

func (p *producer) Terminate() (error){
	//send the Disconnection Packet
	disConPacket := &pb.Packet{
		Type : pb.Type_DIS_CONN_REQUEST,
	}
	err := writePacket(p.conn , disConPacket)

	if err != nil {
		return err
	}

	p.StopListening()
	p.conn.Close()
	return nil
}

func (p* producer) StopListening(){
	if p.listening{
		close(p.stop)
	}
	p.listening = false
}

