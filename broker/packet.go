package main

import (
	"fmt"
	"net"
	"strings"
)

type Packet struct{
	Conn net.Conn
	Type uint8
	Topic string 
	Payload string
}

func newPacket(packetbuffer [2049]byte , conn net.Conn) Packet{
	topicStr := string(packetbuffer[1:2025])	
	payloadStr := string(packetbuffer[2025:2049])

	strings.Trim(topicStr,"\x00")
	strings.Trim(payloadStr,"\x00") //trim the null bytes

	newpacket := Packet{
		Conn: conn,
		Type: uint8(packetbuffer[0]),
		Topic : topicStr,
		Payload :payloadStr,
	}
	return newpacket;
}

func (p *Packet) Print() {
	fmt.Printf("Type: %d\n", p.Type)
	fmt.Printf("Topic: %s\n", p.Topic)
	fmt.Printf("Payload: %s\n", p.Payload)
}

func (p *Packet) toBytes() []byte {
	buf :=  make([]byte , 2049)
	buf[0] = p.Type
	copy(buf[1:2025] , []byte(p.Topic))
	copy(buf[2025:2049] , []byte(p.Topic))
	return buf
}
