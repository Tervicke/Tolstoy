package broker

import (
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

type packetHandler func(packet Packet) bool

var handlers = map[uint8]packetHandler{
	4:HandlePublishPacket, //handles the publish packet
	5:HandleSubscribePacket, //handles the subscribe packet
	6:HandleUnsubscribePacket,
	7:HandleConnectionRequestPacket,
}
func HandleUnsubscribePacket(packet Packet) bool{
	topicConnections , exists := Topics[packet.Topic]
	if !exists{
		return true
	}
	delete(topicConnections , packet.Conn)
	packet.acknowledge() //subscribe ack
	return true;//since no error
}

func HandlePublishPacket(packet Packet) bool{
	clients , exists := Topics[packet.Topic]
	//if the topic doesnt exist create it 
	if !exists{
		Topics[packet.Topic] = make(map[net.Conn]struct{})
	}
	//write it to the log file 
	WriteMessage(packet.Payload, packet.Topic)
	log.Println("Acknowledged")
	for client := range clients{
		packet.Type = 3 // change the packet type to indicating its a server sent packet
		var emptybytearray [2049]byte;
		packetbytes := packet.toBytes()
		if packetbytes == emptybytearray {
			log.Println("Ignored a empty packet")
			return false
		}
		client.Write( packetbytes[:] )
	}

	packet.acknowledge()
	return true
}

func HandleSubscribePacket(packet Packet) bool {
	_ , exists := Topics[packet.Topic]

	if !exists{
		errorpacket := newErrPacket("Topic does not exist");
		log.Printf("Sent Error packet for no topic %v\n",packet.Topic)
		packet.Conn.Write(errorpacket[:]);
		return false;
	}
	//add subscriber
	Topics[packet.Topic][packet.Conn] = struct{}{}
	log.Printf("New subscriber added to %s | count = %d\n",packet.Topic,len(Topics[packet.Topic]))
	packet.acknowledge() //ack code - 11 for successful subscribe
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
func HandleConnectionRequestPacket(packet Packet) bool{
	_,already_connected := ActiveConnections[packet.Conn]
	if already_connected {
		return false
	}

	//add it to the active connection
	activeconnmutex.Lock()
	ActiveConnections[packet.Conn] = struct{}{}
	activeconnmutex.Unlock()

	log.Println("Verified agent")
	return true
}


