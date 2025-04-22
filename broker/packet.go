package main;

import (
	"fmt"
)

type packet struct{
	Type uint8
	Topic [1024]byte
	Payload [1024]byte
}

func newPacket(packetbuffer [2049]byte) packet{
	newpacket := packet{
		Type: uint8(packetbuffer[0]),
		Topic : [1024]byte(packetbuffer[1:1025]),
		Payload :[1024]byte(packetbuffer[1025:2049]),
	}
	return newpacket;
}

func (p *packet) Print() {
	fmt.Printf("Type: %d\n", p.Type)
	fmt.Printf("Topic: %s\n", p.Topic)
	fmt.Printf("Payload: %s\n", p.Payload)
}
