package broker

import (
	//	"log"
	"log"
	"net"

		"os"
		"strconv"
		"time"
	pb "github.com/Tervicke/Tolstoy/internal/proto"

)

type packetHandler func(packetConn net.Conn , packet *pb.Packet) bool

var handlers = map[pb.Type]packetHandler{
	pb.Type_CONN_REQUEST : handleConnectionRequestPacket,
	pb.Type_SUBSCRIBE : handleSubscribePacket,
	pb.Type_UNSUBSCRIBE : handleUnsubscribePacket,
	pb.Type_PUBLISH : handlePublishPacket,
	pb.Type_DIS_CONN_REQUEST:handleDisconnectionPacket,
}
func handleUnsubscribePacket(packetConn net.Conn , packet *pb.Packet) bool{
	topicConnections , exists := Topics[packet.Topic]
	if !exists{
		return true
	}
	delete(topicConnections , packetConn)
	ackPacket(packetConn, packet)
	return true;//since no error
}

func handlePublishPacket(packetConn net.Conn , packet *pb.Packet) bool{
	clients , exists := Topics[packet.Topic]
	//if the topic doesnt exist create it 
	if !exists{
		Topics[packet.Topic] = make(map[net.Conn]bool)
	}
	//write it to the log file 
	WriteMessage(packet.Payload, packet.Topic)
	ackPacket(packetConn , packet)
	for client := range clients{
		deliverPacket := &pb.Packet{
			Type : pb.Type_DELIVER,
			Topic : packet.Topic,
			Payload: packet.Payload,
		}
		writePacket(client,deliverPacket)
	}

	return true
}

func handleSubscribePacket(packetConn net.Conn , packet *pb.Packet) bool {
	_ , exists := Topics[packet.Topic]

	if !exists{
		//make a new Error packet
		errorPacket := packet	
		errorPacket.Type = pb.Type_ERROR
		errorPacket.Error = &pb.ErrorMsg{
			Code:1, //default code for now
			Text:"No such topic exists",
		}
		err := writePacket(packetConn , errorPacket)
		if err != nil {
			log.Println("Failed to send error packet")
		}
		return false;
	}
	//add subscriber
	Topics[packet.Topic][packetConn] = true
	log.Printf("New subscriber added to %s | count = %d\n",packet.Topic,len(Topics[packet.Topic]))
	ackPacket(packetConn , packet)
	return true
}

func WriteMessage(payload string , topic_name string){
	if !brokerSettings.Persistence.Enabled {
		return
	}
	filepath := GetFilePath(time.Now() , topic_name , brokerSettings.Persistence.Directory)
	Writetofile(payload , filepath)

}
func Writetofile(payload , filename string ){
	file,err := os.OpenFile(filename , os.O_APPEND|os.O_CREATE|os.O_WRONLY , 0644)
	if err != nil{
		log.Printf("Error writing payload to log file %v",err)
	}

	_,err = file.Write([]byte(payload + "\n"))

	if err != nil{
		log.Println("Error writing to a file")
	}
}
func GetFilePath(t time.Time , topic_name, persistance_directory string) string {
	year , month , day := t.Date()
	date := strconv.Itoa(year) + "-" + month.String() + "-"  +  strconv.Itoa(day);
	return (persistance_directory + date + "-" + topic_name + ".log")
}

func handleConnectionRequestPacket(packetConn net.Conn , packet *pb.Packet) bool{
	_,already_connected := ActiveConnections[packetConn]
	if !already_connected {
		activeconnmutex.Lock()
		ActiveConnections[packetConn] = struct{}{}
		activeconnmutex.Unlock()
	}
	log.Println("Verified agent")
	return true
}

func handlePacket(packetConn net.Conn , packet *pb.Packet){
	_ , verified := ActiveConnections[packetConn] 
	if !verified && packet.Type != pb.Type_CONN_REQUEST {
		log.Println("Unverfied packet recieved ... ignored")
		return
	}

	if function , exists := handlers[packet.Type] ; exists {
		result := function(packetConn , packet)
		if !result {
			log.Println("Packet not cleared")
		}else{
			//write ack packet
			log.Println("Packet cleared")
			ackpacket := &pb.Packet{
				Type: pb.Type_ACK_CONN_REQUEST,
			}
			writePacket(packetConn , ackpacket)
		}
	}else{
		log.Println("Invalid packet type recieved")
	}
}
func handleDisconnectionPacket(packetConn net.Conn , packet *pb.Packet) bool {
	activeconnmutex.Lock()
	delete(ActiveConnections , packetConn)
	activeconnmutex.Unlock()
	ackPacket(packetConn , packet)
	return true
}
