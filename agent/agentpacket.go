package agent

import (
	"strings"
)

type Packet struct{
	Type uint8
	Topic string
	Payload string
}

func newpacket(packetbuffer [2049]byte ) Packet{
	topicStr := string(packetbuffer[1:1025])        
	payloadStr := string(packetbuffer[1025:2049])  


	strings.Trim(topicStr,"\x00")
	strings.Trim(payloadStr,"\x00") //trim the null bytes

	newpacket := Packet{
		Type: uint8(packetbuffer[0]),
		Topic : topicStr,
		Payload :payloadStr,
	}
	return newpacket;
}

func (p* Packet) tobytes() []byte{

	buf := make([]byte,2049)
	buf[0] = p.Type
	copy(buf[1:1025] , []byte(p.Topic) )
	copy(buf[1025:2049] , []byte(p.Payload) )
	return buf[:]
}

func makepacket(Type int , Topic string , Payload string) Packet{
	newpacket := Packet{
		Type: uint8(Type),
		Topic:(Topic),
		Payload:(Payload),
	}
	return newpacket
}
