package broker

import (
	pb "Tolstoy/proto"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"

	//	"os"
	//	"os/signal"
	"strconv"
	"sync"

	//	"syscall"

	"google.golang.org/protobuf/proto"
)

//make a map of net Conn and struct (because 0 bytes) //simulate a set

var (
	ActiveConnections = make(map[net.Conn]struct{})
	activeconnmutex sync.Mutex
)

var Topics = make(map[string]map[net.Conn]struct{})

//host and port 
var Host = "localhost"
var Port = "8080"

type configdata struct{
		Port int  `yaml:"Port"`
		Host string `yaml:"Host"`
		Topics []string `yaml:"Topics"`
		Persistence struct{ 
			Enabled bool `yaml:"Enabled"`
			Directory string `yaml:"Directory"`
		} `yaml:"Persistence"`
}
//default data to be overwritten when loadconfig runs
var brokerSettings = configdata{
	Port: -1,
	Host:"",
};

func handleConnection(curConn net.Conn){
	defer curConn.Close()
	//infinite loop to keep listening
	fixedSize := 4;
	bufSize := make([]byte , fixedSize)
	for {
		_ , err := io.ReadFull(curConn , bufSize)
		if err != nil {
			log.Println("Connection lost")
			return
		}
		msgLen := binary.BigEndian.Uint32(bufSize)

		if msgLen > 10_000 {
			log.Println("message len limit excedded closing connection")
			return
		}

		msgBuf := make([]byte , msgLen)
		_ , err = io.ReadFull(curConn, msgBuf)
		if err != nil {
			log.Println("Connection lost")
			return
		}
		var recPacket =  &pb.Packet{}
		err = proto.Unmarshal(msgBuf , recPacket)
		if err != nil {
			log.Println("Connection lost")
			continue
		}
		handlePacket(curConn , recPacket)
		if recPacket.Type == pb.Type_ACK_DIS_CONN_REQUEST {
			return
		}
	}
}

func StartServer(configpath string){
	go func(){
		log.Println("Pprof listening at http://localhost:6060/debug/pprof/")
		log.Println(http.ListenAndServe("localhost:6060",nil))
	}()

	//handleCrash()

	err := loadConfig(configpath)

	if err != nil{
		log.Panicf("Error occured %v",err)
	}

	addr := brokerSettings.Host + ":" + strconv.Itoa(brokerSettings.Port)

	broker , err := net.Listen("tcp" , addr);

	if err != nil{
		log.Panicf("Failed to start the broker %v \n",err)
	}

	log.Printf("Server started on %s \n",addr)

	for true {
		conn , err := broker.Accept();
		if err != nil {
			fmt.Println("failed to accept a request");
			fmt.Println(err);
		}

		go handleConnection(conn);
	}
}

/*
func handleCrash(){

	c := make(chan os.Signal , 1)
	signal.Notify(c,os.Interrupt,syscall.SIGTERM)
	go func(){
		 <-c
		log.Printf("Shutting Down..sending disconnection packets to all the %d agents",len(ActiveConnections))
		dpacket := newDisconnectionPacket()
		for conn := range ActiveConnections{
			conn.Write(dpacket.toBytes())
		}
		log.Printf("Sent disconnection packets to all the %d agents",len(ActiveConnections))
		os.Exit(0)
	}()
}
*/
