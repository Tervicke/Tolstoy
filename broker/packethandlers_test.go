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
			expected_topic = "test-topic"
			expected_payload = "test-payload will get ignored"
		)

		var buf [2049]byte;
		buf[0] = 5
		copy(buf[1:1025],expected_topic)
		copy(buf[1025:2049],expected_payload)
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
		_,contains := Topics[expected_topic][subpacket.Conn]
		if contains != true{
			t.Errorf("contains expected %t got %t",true,contains)
		}
		ack := <-ackCh
		expectedack := makepacketbyte(11,expected_topic,expected_payload)
		testAck(ack,expectedack,t)
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
	const (
		expected_topic = "test topic"
		expected_payload = "test payload"
	)
	unsubbytes:= makepacketbyte(6,expected_topic,expected_payload)

	conn1, conn2 := net.Pipe()

	defer conn1.Close()
	defer conn2.Close()

	ackCh := make(chan [2049]byte)
	//read ack and pass it to ack channel	
	go func(){
		var ack [2049]byte
		_,err := conn2.Read(ack[:])
		if err != nil{
			t.Errorf("Failed to read ack")
		}
		ackCh<-ack
	}()

	unsubpacket := newPacket(unsubbytes,conn1)
	//check if the function handler works for the unsubpacket
	handlerfunction , exists :=  handlers[unsubpacket.Type]

	if !exists{
		t.Fatal("handler function expected for the unsubpacket type")
	}
	//make the topic and add the net connection to the list of topics first 	
	Topics[unsubpacket.Topic] = make(map[net.Conn]struct{})
	Topics[unsubpacket.Topic][unsubpacket.Conn] = struct{}{}


	_ , exists  = Topics[unsubpacket.Topic][unsubpacket.Conn]

	if !exists {
		t.Error("Error when trying to add connection to the topics")
	}

	//action
	response := handlerfunction(unsubpacket)
	if !response {
		t.Errorf("response not true for valid unsubpacket")
	}
	_ , exists = Topics[unsubpacket.Topic][unsubpacket.Conn]

	if exists {
		t.Errorf("Conn not deleted ")
	}
	ack := <-ackCh
	expectedack := makepacketbyte(12,expected_topic,expected_payload)
	testAck(ack,expectedack,t)
}

func TestPacketHandlerWithPublishePacket(t *testing.T){
	const (
		expected_topic = "hello-topic"
		expected_payload = "hello-payload"
	)
	pubbyte:= makepacketbyte(4,expected_topic,expected_payload)
	conn1,conn2 := net.Pipe()
	defer conn1.Close()
	defer conn2.Close()

	ackCh := make(chan [2049]byte)
	go func(){
		var ack [2049]byte
		_,err := conn2.Read(ack[:])
		if err != nil {
			t.Errorf("failed to read pubish ack")
		}
		ackCh<-ack
	}()

	pubpacket := newPacket(pubbyte , conn1)
	//test handler functions exists for pubpacket
	handlerfunction , exists := handlers[pubpacket.Type]
	if !exists{
		t.Fatalf("Handler function expected did not find")
	}
	//action
	sent := handlerfunction(pubpacket)
	if !sent{
		t.Errorf("Handlerfunction returns false expected true")
	}
	ack := <-ackCh

	expectedack := makepacketbyte(10,expected_topic,expected_payload)

	testAck(ack,expectedack,t)
}

//helper function to test ack
func testAck(actualack , expectedack  [2049]byte , t *testing.T){

	actual_type := actualack[0]
	expected_type := expectedack[0]
	if actual_type != expected_type{
		t.Errorf("ack type mismatch expected %d got %d",expected_type , actual_type)
	} 

	actual_topic := strings.Trim(string(actualack[1:1025]),"\x00")
	expected_topic := strings.Trim(string(expectedack[1:1025]),"\x00")
	if actual_topic != expected_topic {
		t.Errorf("ack topic mismatch expected %s got %s",expected_topic,actual_topic)
	} 

	actual_payload := strings.Trim(string(actualack[1025:2049]),"\x00")
	expected_payload:= strings.Trim(string(expectedack[1025:2049]),"\x00")
	if actual_payload != expected_payload{
		t.Errorf("ack payload mismatch expected %s got %s",expected_payload,actual_payload)

	} 
}
