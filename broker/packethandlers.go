package broker

import (
	//	"log"
	"log"
	"net"

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
	pb.Type_PAUSE:handlePausePacket,
	pb.Type_RESUME:handleResumePacket,
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
	topicname := packet.Topic
	//if the topic doesnt exist create it 
	if !exists{
		Topics[packet.Topic] = make(map[net.Conn]bool)
		Topics[topicname] = make(map[net.Conn]bool) 
		indexFileName := brokerSettings.Persistence.Directory + topicname + ".index"
		lastOffset , err := getLastOffset(indexFileName)
		if err != nil { 
			log.Println("failed to read last offset for the file ",indexFileName)
			return false
		}
		TopicOffsets[topicname] = lastOffset 
	}

	//increase the offset to match the current one
	TopicOffsets[topicname]++
	//write it to the log file 


	//WriteMessage(packet.Payload, topicname)
	//create the record to append it
	now := time.Now()
	record := &pb.Record{
		Timestamp: now.Unix(),
		Payload: []byte(packet.Payload),
	}

	indexFileName := brokerSettings.Persistence.Directory + topicname + ".index"
	logfileName := brokerSettings.Persistence.Directory  + topicname +".log"

	//update the offset in the index file
	err := updateOffset(indexFileName , TopicOffsets[topicname])
	if err != nil {
		log.Println("failed to update offset for ",topicname,err)
	}
	
	//append the record 
	pos , err := AppendRecord(  logfileName, record )

	if err != nil {
		log.Println("Failed to append record" , err)
		return false;
	}
	//append the pos in the index file
	err = appendIndex(indexFileName , pos)
	if err != nil {
		log.Println("failed to append index",err)
	}

	ackPacket(packetConn , packet)

	//delivery packet
	deliverPacket := &pb.Packet{
		Type : pb.Type_DELIVER,
		Topic : packet.Topic,
		Payload: packet.Payload,
		Offset:TopicOffsets[topicname],
	}

	for client,resumed:= range clients{
		//only send the packet if its not paused 
		if resumed{
			writePacket(client,deliverPacket)
		}
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

func GetFilePposath(t time.Time , topic_name, persistance_directory string) string {
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
	ackpacket := &pb.Packet{
		Type: pb.Type_ACK_CONN_REQUEST,
	}
	writePacket(packetConn , ackpacket)
	log.Println("Verified agent")
	return true
}

func handlePacket(packetConn net.Conn , packet *pb.Packet){
	_ , verified := ActiveConnections[packetConn] 
	if !verified && packet.Type != pb.Type_CONN_REQUEST {
		log.Println("Unverfied packet recieved ... ignored")
		return
	}
	if packet.Type == pb.Type_CONN_REQUEST{
		handleConnectionfunction := handlers[pb.Type_CONN_REQUEST]
		handleConnectionfunction(packetConn , packet)
		log.Println("verified user")
		return
	}
	if function , exists := handlers[packet.Type] ; exists {
		result := function(packetConn , packet)
		if !result {
			log.Println("Packet not cleared")
		}else{
			//write ack packet
			log.Println("Packet cleared")
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
func handlePausePacket(packetConn net.Conn , packet *pb.Packet) bool {
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
	Topics[packet.Topic][packetConn] = false; 
	ackPacket( packetConn , packet )
	return true
}
func handleResumePacket(packetConn net.Conn , packet *pb.Packet) bool {
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
	//make it to true to send latest messages
	Topics[packet.Topic][packetConn] = true; 
	ackPacket( packetConn , packet )

	if packet.Offset == -1{ //latest
		return true
	}

	offset := packet.Offset
	indexFileName :=  brokerSettings.Persistence.Directory + packet.Topic + ".index"
	logFileName := brokerSettings.Persistence.Directory + packet.Topic + ".log"
	totalOffset := TopicOffsets[packet.Topic] 
	for offset < totalOffset{
		pos , err := getOffsetPos(indexFileName , offset)
		if err != nil {
			log.Println("Failed to read pos skipping" , err)
			offset++
			continue
		}
		record , err := readFromRecord(logFileName , pos)
		if err != nil {
			log.Println("Failed to read Record skipping ", err)
			offset++
			continue
		}
		deliverPacket := &pb.Packet{
			Type : pb.Type_DELIVER,
			Topic : packet.Topic,
			Payload: string(record.Payload),
			Offset:offset,
		}
		offset++
		writePacket(packetConn , deliverPacket)
	}

	return true
}
