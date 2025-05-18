package agent

import (
	pb "github.com/Tervicke/Tolstoy/internal/proto"

	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"net"
	"os"

	"google.golang.org/protobuf/proto"
)

type agent struct{
	conn net.Conn
	stop chan struct{}
	listening bool //true - listening channel exists
	ackchan chan Packet //publish ack channel
	callbacks map[string]OnMessage //map for the callbacks of various topics
}

type OnMessage func(topic  , message string)

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
