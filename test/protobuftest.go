package main

import (
	pb "Tolstoy/proto"
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"google.golang.org/protobuf/proto"
)

func testproto() {
	p := &pb.Packet{
		Type		: pb.Type_SUBSCRIBE,
		Topic		: "mytopic",
		Payload	: "my payload",
	}
	fmt.Printf("packet is %v+",p)
	data , err := proto.Marshal(p)
	if err != nil {
		log.Fatal("proto marshal failed")
	}
	conn , err := net.Dial("tcp" , "localhost:8080")
	if err != nil {
		log.Fatal("failed to dial")
	}
	size := make([]byte,4)
	binary.BigEndian.PutUint32(size , uint32(len(data)))
	conn.Write(size)
	conn.Write(data)
}
