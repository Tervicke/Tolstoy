package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"gopkg.in/yaml.v3"
)

//make a map of net Conn and struct (because 0 bytes) //simulate a set
var ActiveConnections = make(map[net.Conn]struct{})
var Topics = make(map[string]map[net.Conn]struct{})

//host and port 
var Host = "localhost"
var Port = "8080"

func handleConnection(curCon net.Conn){
	log.Println("New Agent joined")
	defer curCon.Close()
	ActiveConnections[curCon] = struct{}{}
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
		if handlepacket, ok := handlers[newpacket.Type]; ok {
			log.Println("packet recieved....acknowledging and handling")
			if handlepacket(newpacket) {
				log.Println("handled")
				//newpacket.acknowledge()
				//fmt.Println("acknowledged")
			}else{
				log.Println("Could not acknowledge , error")
			}
		}else{
			//specify error code and and send it accordingly 
			log.Println("Recieved Invalid packet type")
		}
	}
}

func main() {
	handleCrash()

	config_path := flag.String("config" , "config.yaml" ,"The path to the config.yaml file \n by default it searches the current directory")                              
	flag.Parse()
	
	//check if the config_path exists

	loadConfig(*config_path)

	addr := Host + ":" + Port

	broker , err := net.Listen("tcp" , addr);

	if err != nil{
		log.Panicf("Failed to start the broker %v \n",err)
	}

	log.Printf("Server started on %s \n",addr)

	//server starts and then the config is loaded 

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

func loadConfig(config_path string) () {
	//check if the file exists
	_ , err := os.Stat(config_path)
	if os.IsNotExist(err){
		log.Panicln("Couldnt find the config file")
	}

	//parse the config

	type configdata struct {
		Port int  `yaml:"Port"`
		Host string `yaml:"Host"`
		Topics []string `yaml:"Topics"`
	} 

	var config configdata
	config_file,_ := os.ReadFile(config_path)

	err = yaml.Unmarshal( config_file , &config)

	if err != nil{
		log.Panicf("Failed to parse the error %v",err)
	}

	//set the port and host
	Host = config.Host
	Port = strconv.Itoa(config.Port)

	//add the predefine the topics
	for _,topicname := range config.Topics{
		Topics[topicname] = make(map[net.Conn]struct{}) 
	}

	log.Println("Loaded config succesfully")
}
