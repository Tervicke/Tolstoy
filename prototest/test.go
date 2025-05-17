package main

import (
	pb "Tolstoy/proto"
	"encoding/binary"
	"io"
	"log"
	"net"
	"fmt"
	"google.golang.org/protobuf/proto"
)

func main() {
	//runsub()
	connPacket := &pb.Packet{
		Type		: pb.Type_CONN_REQUEST,
	}
	conn , err := net.Dial("tcp" , "localhost:8080")
	defer conn.Close()
	if err != nil {
		log.Fatal("failed to dial")
	}
	writePacket(conn , connPacket)

	pubPacket := &pb.Packet{
		Type		: pb.Type_PUBLISH,
		Topic		: "mytopic",
		Payload : "unsubscribe",
	}
	writePacket(conn , pubPacket)
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
	/*
	size := make([]byte,4)
	binary.BigEndian.PutUint32(size , uint32(len(data)))
	fmt.Printf("%+v\n",p)
	fmt.Println(len(data))
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
	fmt.Printf("%+v\n",packet)
 	for {

	}
	*/
	fmt.Println("send the packet successfully")
}

