package agent

import (
	"strings"
)

type packet struct{
	Type uint8
	Topic string
	Payload string
}

func newpacket(packetbuffer [2049]byte ) packet{
	topicStr := string(packetbuffer[1:1025])        
	payloadStr := string(packetbuffer[1025:2049])  


	strings.Trim(topicStr,"\x00")
	strings.Trim(payloadStr,"\x00") //trim the null bytes

	newpacket := packet{
		Type: uint8(packetbuffer[0]),
		Topic : topicStr,
		Payload :payloadStr,
	}
	return newpacket;
}

