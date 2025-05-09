package broker

import (
	"net"
	"testing"
)

type testcase struct{
	name string 
	expectedtype uint8
	expectedtopic string
	expectedpayload string
}

func TestNewPacket(t *testing.T){
	
	tests := []testcase{
		{
			name : "test all filled input fields",
			expectedtype : 1,
			expectedtopic : "Test Topic" ,
			expectedpayload : "Test Payload",
		},
		{
			name : "test empty topic",
			expectedtype : 1,
			expectedtopic : "",
			expectedpayload : "Test Payload",
		},
		{
			name : "test empty payload",
			expectedtype : 1,
			expectedtopic : "Test Topic",
			expectedpayload : "", //empty payload
		},
	}
	
	for _,tt:= range tests {
		var buf [2049]byte 
		buf[0] = tt.expectedtype
		copy(buf[1:1025],tt.expectedtopic)
		copy(buf[1025:2049],tt.expectedpayload)
		dummyconnection,_ := net.Pipe()
		t.Run(tt.name , func(t *testing.T){
			actual := newPacket(buf,dummyconnection)
			if actual.Type != tt.expectedtype{
				t.Errorf("Expected Type %d got %d",tt.expectedtype, actual.Type)
			}
			if actual.Topic != tt.expectedtopic{
				t.Errorf("Expected Topic %s got %s",tt.expectedtopic, actual.Topic)
			}
			if actual.Payload != tt.expectedpayload{
				t.Errorf("Expected Payload %s got %s",tt.expectedpayload, actual.Payload)
			} 
		})
	}
}

func TestToBytes(t *testing.T) {

	tests := []testcase{
		{
			name : "test all filled input fields",
			expectedtype : 1,
			expectedtopic : "Test Topic" ,
			expectedpayload : "Test Payload",
		},
		{
			name : "test empty topic",
			expectedtype : 1,
			expectedtopic : "",
			expectedpayload : "Test Payload",
		},
		{
			name : "test empty payload",
			expectedtype : 1,
			expectedtopic : "Test Topic",
			expectedpayload : "",
		},
		{
			name : "test everything empty",
			expectedtopic : "",
			expectedpayload : "", 
		},
	}		
	for _,tt := range tests{
		var exptectedbuf[2049]byte
		exptectedbuf[0] = tt.expectedtype
		copy(exptectedbuf[1:1025],tt.expectedtopic)
		copy(exptectedbuf[1025:2049],tt.expectedpayload)
		dummyconnection,_ := net.Pipe()
		newpacket := newPacket(exptectedbuf,dummyconnection)
		t.Run(tt.name , func(t *testing.T) {
			actualbuf := newpacket.toBytes()
			if actualbuf != exptectedbuf {
				t.Errorf("Expected buf was different than the actualbuf")
			}
		})
	}
}

func TestNewErrPacket(t *testing.T){
	type testcase struct{
		name string
		expectedtype uint8
		expected string
	}
	tests := []testcase{
		{
			name: "test invalid topic",
			expectedtype: 2,
			expected: "Topic does not exist",
		},
	} 
	for _,tt := range tests{
		t.Run(tt.name , func(t *testing.T){
			var expectedbuf [2049]byte;
			expectedbuf[0] = tt.expectedtype
			copy(expectedbuf[1:],tt.expected)
			actualbuf := newErrPacket(tt.expected)
			if actualbuf != expectedbuf{
				t.Errorf("Expected Error packet not got")
			}
		})
	}
}

func TestAcknowledge(t *testing.T){
	type testcase struct{
		name string
		packetype uint8
		expectedtype uint8
		expectedtopic string
		expectedpayload string 
	}
	tests := []testcase{
		{
			name : "test publish ack",
			packetype: 4,
			expectedtype: 10,
			expectedtopic: "test topic",
			expectedpayload: "publish payload",
		},
		{
			name : "test subscribe ack",
			packetype: 5,
			expectedtype: 11,
			expectedtopic: "test topic",
			expectedpayload: "subscribe payload",
		},
		{
			name : "test unsubscribe ack",
			packetype: 6,
			expectedtype: 12,
			expectedtopic: "test topic",
			expectedpayload: "unsubscribe payload",
		},
	}
	packetconn , verifyconn := 	net.Pipe()

	defer packetconn.Close()
	defer verifyconn.Close() 

	for _,tt := range tests{
		tt := tt

		t.Run(tt.name , func(t *testing.T) {
			var expectedbuf [2049]byte
			expectedbuf[0] = tt.packetype
			copy(expectedbuf[1:1025],tt.expectedtopic)
			copy(expectedbuf[1:1025],tt.expectedpayload)
			newpacket := newPacket(expectedbuf,packetconn)
			go newpacket.acknowledge()
			var actualbuf [2049]byte;
			_,err := verifyconn.Read(actualbuf[:])

			if err != nil{
				t.Errorf("error reading the ack packet")
			}

			//update the expected buf to change the expected type
			expectedbuf[0] = tt.expectedtype

			if actualbuf != expectedbuf {
				t.Errorf("Did not recieve exact acknowledge packet")
			}

		})
	}
	
}
