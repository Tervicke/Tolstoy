package agent

import (
	pb "github.com/Tervicke/Tolstoy/internal/proto"

	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"math/rand/v2"
	"errors"
	"io"
	"time"
	"syscall"

	"google.golang.org/protobuf/proto"
)

type listener interface{
	listen()
}

type agent struct {
	conn        net.Conn
	stop        chan struct{}
	listening   bool                       //true - listening channel exists
	callbacks   map[string]OnMessage       //map for the callbacks of various topics
	ackchannels map[string]chan *pb.Packet //channel for handling ack
	MaxAttempts int
	serverAddr  string      //addr of the connected server
	tlsCfg      *tls.Config //tls config if any nil if none
	listener listener //the listener interface to call the listen()
}

func (a *agent) listen() {
	for {
		select {
		case <-a.stop:
			return
		default:
			//read the size first and then read the buf and deserialize it
			sizeBuf := make([]byte, 4)
			io.ReadFull(a.conn, sizeBuf)
			size := binary.BigEndian.Uint32(sizeBuf)
			msgBuf := make([]byte, size)
			io.ReadFull(a.conn, msgBuf)
			//serialize
			packet := &pb.Packet{}
			proto.Unmarshal(msgBuf, packet)
			switch packet.Type {
			default:
				a.ackchannels[packet.RequestId] <- packet
			}
		}
	}
}

//Used to terminate the producer
func (a *agent) Terminate() error {
	//send the Disconnection Packet
	disConPacket := &pb.Packet{
		Type: pb.Type_DIS_CONN_REQUEST,
	}
	err := a.safeWritePacket(disConPacket) 
	if err != nil {
		return err
	}

	a.stoplistening()
	a.conn.Close()
	return nil
}

//Internal used to stop listening to the server and terminate the listen go routine
func (a *agent) stoplistening() {
	if a.listening {
		close(a.stop)
	}
	a.listening = false
}


// refers to the broken pipe returns true if broken pipe fixed after certain attemps else returns false
func (a *agent) brokenPipe() bool {
	for i := 1; i <= a.MaxAttempts; i++ {
		var err error = nil
		var conn net.Conn
		if a.tlsCfg != nil {
			conn, err = tls.Dial("tcp", a.serverAddr, a.tlsCfg)
		} else {
			//tls config
			conn, err = net.Dial("tcp", a.serverAddr)
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
		a.conn = conn
		go a.listen()
		return true
	}
	return false
}

//Internal safe write packet used to check and fix error when writing packet
func (a *agent) safeWritePacket(packet *pb.Packet) (error){
	for i := 1; i <= a.MaxAttempts; i++ {
		err := writePacket(a.conn, packet)
	
		if err != nil {
			//reconnect and try to send the message then
			if errors.Is(err, syscall.EPIPE) {
				fixed := a.brokenPipe()
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

type OnMessage func(topic string  , payload []byte)

func writePacket(packetConn net.Conn , packet *pb.Packet) (error) {
	data , err := proto.Marshal(packet)
	if err != nil {
		return err 
	}
	size := make([]byte , 4)
	binary.BigEndian.PutUint32(size , uint32(len(data)))
	_,err = packetConn.Write(size)

	if err != nil {
		return err
	}

	_,err = packetConn.Write(data)

	if err != nil {
		return err
	}
	return nil
}

func LoadTLSConfig(certFile string) (*tls.Config , error) {
	caCert , err := os.ReadFile(certFile)
	if err != nil {
		return nil,err
	}
	//create a CA certificate pool and add cert
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert){
		return nil,fmt.Errorf("failed to open a CA cert")
	}
	return &tls.Config{
		RootCAs: caCertPool,
		InsecureSkipVerify: false,
	},nil
}

//generates a unique Id by calling the generateId function and also verifying the existing id
func generateUniqueId(channels map[string]chan *pb.Packet) (string) {
	for {
		Id := generateId()
		_ , exist := channels[Id] 
		if !exist {
			return Id;
		}
	}
}
//generates a random id from a const letters
func generateId() string{
	const letters string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	var Id string = "";
	for  i := 0 ; i <= 8 ; i++ {
		Id += string(letters[rand.IntN(len(letters))]) //generate a random character from the string
	}
	return Id;
}
