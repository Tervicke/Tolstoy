package broker

import (
	"net"
	"strings"
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
	//Test with valid topic 
	t.Run("test subpacket with a valid topic" , func(t *testing.T) {

		const (
			topic = "test-topic"
			payload = "test-payload will get ignored"
		)

		var buf [2049]byte;
		buf[0] = 5
		copy(buf[1:1025],topic)
		copy(buf[1025:2049],payload)
		conn1,conn2  := net.Pipe()  
		defer conn2.Close()
		defer conn1.Close()
		subpacket := newPacket(buf, conn1)
		
		//make the ack and handle it
		ackCh := make(chan [2049]byte)

		go func(){
			var ack [2049]byte
			_,err := conn2.Read(ack[:])
			if err != nil{
				t.Errorf("error when reading ack")
			}
			ackCh <- ack
		}()

		//make sure the Topic exists
		Topics[subpacket.Topic] =  make(map[net.Conn]struct{})
		handlerfunction , exists := handlers[subpacket.Type] 
		
		//make sure handlerfunction exists for sub packet
		if !exists{
			t.Fatal("expected handlerfunction for subscribe packet")
		}

		//action
		added := handlerfunction(subpacket)
		if added != true{
			t.Errorf("Expected added to be %t found %t",true,added)
		}
		//make sure the connection exists
		_,contains := Topics[topic][subpacket.Conn]
		if contains != true{
			t.Errorf("contains expected %t got %t",true,contains)
		}
		ack := <-ackCh
		if ack[0]  != 11{
			t.Errorf("Incorrect ack type recieved expected %d got %d",10,ack[0])
		}
		actual_topic := strings.Trim(string(ack[1:1025]),"\x00")
		if  actual_topic != topic {
			t.Errorf("Topic mismatch E:%s G:%s",topic,actual_topic)
		}
		actual_payload := strings.Trim(string(ack[1025:2049]),"\x00")
		if actual_payload != payload {
			t.Errorf("Payload mismatch E:%s G:%s",payload,actual_payload)
		}

	})

	//test with invalid topic
	t.Run("test subpacket with a invalid topic" , func(t *testing.T) {

		const (
			topic = "invalid-topic"
			payload = "test-payload will get ignored"
		)

		var buf [2049]byte;
		buf[0] = 5
		copy(buf[1:1025],topic)
		copy(buf[1025:2049],payload)
		conn1,conn2  := net.Pipe()  
		defer conn2.Close()
		defer conn1.Close()
		subpacket := newPacket(buf, conn1)
		
		//make the ack and handle it
		ackCh := make(chan [2049]byte)

		go func(){
			var ack [2049]byte
			_,err := conn2.Read(ack[:])
			if err != nil{
				t.Errorf("error when reading ack")
			}
			ackCh <- ack
		}()

		//make sure the Topic does not exist
		//Topics[subpacket.Topic] =  make(map[net.Conn]struct{})
		handlerfunction , exists := handlers[subpacket.Type] 
		
		//make sure handlerfunction exists for sub packet
		if !exists{
			t.Fatal("expected handlerfunction for subscribe packet")
		}

		//action
		added := handlerfunction(subpacket)
		if added != false{
			t.Errorf("Unexpected behaviour %t found %t",true,added)
		}
		//make sure the connection does not exist
		_,contains := Topics[topic][subpacket.Conn]
		if contains != false{
			t.Errorf("did not expect to contain %t got %t",true,contains)
		}
		ack := <-ackCh
		//ack should be an error packet
		errorpacket := newErrPacket("Topic does not exist");
		if errorpacket != ack{
			t.Errorf("Expected error packet did not found")
		}
	})
}

func TestPacketHandlerWithUnsubscribePacket(t *testing.T){
	//TODO
}

func TestPacketHandlerWithPublishePacket(t *testing.T){
	//TODO
}
