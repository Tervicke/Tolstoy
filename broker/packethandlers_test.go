package broker

import (
	"net"
	"testing"
)

func TestPacketHandlerWithInvalidPacket(t *testing.T){
	t.Run("test packet handler with invalid packet type" , func(t *testing.T) {
		var buf [2049]byte;
		buf[0] = 10
		copy(buf[1:1025],"expected topic")
		conn1,_  := net.Pipe()  
		invalidpacket := newPacket(buf,conn1)
		expected := false;
		if _,actual:= handlers[invalidpacket.Type] ; actual != expected {
			t.Errorf("Expected false on passing invalid packet type")
		}
	
	})

}

func TestPacketHandlerWithSubscribePacket(t *testing.T){
	//TODO
}

func TestPacketHandlerWithUnsubscribePacket(t *testing.T){
	//TODO
}

func TestPacketHandlerWithPublishePacket(t *testing.T){
	//TODO
}
