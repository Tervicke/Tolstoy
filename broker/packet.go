package main

import (
	"fmt"
	"net"
	"strings"
)
//this is the Acknowledge packet sent by the broker when the message / request is handled is handled
//the topic and the payload will be same of whatever was sent to the server
type AckPacket struct{
	Type uint8
	Topic string
	Payload string
}

//This is the Err packet that is sent by the broker when the handler finds and error in the request / message 
//Error will contain the error as string and Type will be a error type
type ErrPacket struct{
	Type uint8
	Error string	
}

type Packet struct{
	Conn net.Conn
	Type uint8
	Topic string 
	Payload string
}

func newPacket(packetbuffer [2049]byte , conn net.Conn) Packet{
	topicStr := string(packetbuffer[1:1025])        
	payloadStr := string(packetbuffer[1025:2049])  


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
	copy(buf[1025:2049] , []byte(p.Payload))
	return buf
}
func newErrPacket(err string) [2049]byte{
	var errpacket [2049]byte;
	errpacket[0] = 2;
	copy(errpacket[1:], []byte(err))
	return errpacket
}
