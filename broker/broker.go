package broker

import (
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

//make a map of net Conn and struct (because 0 bytes) //simulate a set
var ActiveConnections = make(map[net.Conn]struct{})
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

func handleConnection(curCon net.Conn){
	defer curCon.Close()
	for {
		buf := make([]byte,2049)
		totalread := 0;
		for totalread < 2049 {
			n , err := curCon.Read(buf[totalread:])
			if err != nil{
				fmt.Println("agent left")
				delete(ActiveConnections ,curCon)
				return;
			}
			totalread += n
		}
		
		newpacket := newPacket([2049]byte(buf),curCon);
		handlepacket , ok := handlers[newpacket.Type]
		_,valid_connection := ActiveConnections[curCon] 

		//either its a valid connection or its trying to make it to a connection if not log and return

		if !valid_connection{
			if newpacket.Type != 7{
				log.Println("Packet recieved by unverified connection")
				errorpacket := newErrPacket("Please verify yourself");
				log.Println(newpacket.Type , newpacket.Topic, newpacket.Payload)
				curCon.Write(errorpacket[:])
				return
			}
		}

		if ok {

			log.Println("packet recieved....trying to acknowledge and handle")

			//handle the packet 
			handled := handlepacket(newpacket)

			if handled {
				log.Println("handled the package")
			}else{
				log.Println("Could not handle the package")
			}

		}else{

			//specify error code and and send it accordingly 
			log.Println("Recieved Invalid packet type")

		}

	}
}

func StartServer(configpath string){
	go func(){
		log.Println("Pprof listening at http://localhost:6060/debug/pprof/")
		log.Println(http.ListenAndServe("localhost:6060",nil))
	}()

	handleCrash()

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

