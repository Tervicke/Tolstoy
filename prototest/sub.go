package main

import (
	pb "Tolstoy/proto"
	"encoding/binary"
	"io"
	"log"
	"net"

	"google.golang.org/protobuf/proto"
)

func runsub() {
	p := &pb.Packet{
		Type		: pb.Type_CONN_REQUEST,
	}
	data , err := proto.Marshal(p)
	if err != nil {
		log.Fatal("proto marshal failed")
	}
	conn , err := net.Dial("tcp" , "localhost:8080")
	defer conn.Close()
	if err != nil {
		log.Fatal("failed to dial")
	}
	size := make([]byte,4)
	binary.BigEndian.PutUint32(size , uint32(len(data)))
	conn.Write(size)
	conn.Write(data)
	p = &pb.Packet{
		Type		: pb.Type_SUBSCRIBE,
		Topic : "mytopic",
		Payload : "helloworld",
	}
	data , err = proto.Marshal(p)
	if err != nil {
		log.Fatal("proto marshal failed")
	}
	binary.BigEndian.PutUint32(size , uint32(len(data)))
	conn.Write(size)
	conn.Write(data)
	//read the ack first the size and then the actual packet
	sizebuf := make([]byte , 4)
	_ , err = io.ReadFull(conn,sizebuf)
	if err != nil {
		log.Println("Eror reading the size")
	}
	msgLen := binary.BigEndian.Uint32(sizebuf)
	msgBuf := make([]byte,msgLen) 
	_,err = io.ReadFull(conn,msgBuf)
	if err != nil {
		log.Println("Error when reading the msg")
	}
	//deserialize
	packet := &pb.Packet{}
	err = proto.Unmarshal(msgBuf , packet)
	if err != nil {
		log.Println("failed to unmarshal the message")
	}
	if packet.Type == pb.Type_ERROR {
		log.Fatal("Did not subscribe")
	}else{
		log.Println("Subscribed success")
	}
	unsubPacket := &pb.Packet{
		Type: pb.Type_UNSUBSCRIBE,
		Topic : "mytopic",
	}
	writePacket(conn , unsubPacket)
 	for {

	}
}

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
