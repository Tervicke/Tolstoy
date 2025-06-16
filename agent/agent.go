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

	"google.golang.org/protobuf/proto"
)

type agent struct{
	conn net.Conn
	stop chan struct{}
	listening bool //true - listening channel exists
	ackchan chan Packet //publish ack channel
	callbacks map[string]OnMessage //map for the callbacks of various topics
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
//generates a unique Id
func generateUniqueId(channels map[string]chan *pb.Packet) (string) {
	for {
		Id := generateId()
		_ , exist := channels[Id] 
		if !exist {
			return Id;
		}
	}
}

func generateId() string{
	const letters string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	var Id string = "";
	for  i := 0 ; i <= 8 ; i++ {
		Id += string(letters[rand.IntN(len(letters))]) //generate a random character from the string
	}
	return Id;
}
