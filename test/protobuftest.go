package main

import (
	pb "Tolstoy/proto"
	"fmt"
)

func testproto() {
	p := &pb.Packet{
		Type		: pb.Type_PUBLISH,
		Topic		: "mytopic",
		Payload	: "my payload",
	}
	fmt.Printf("packet is %v+",p)
}
