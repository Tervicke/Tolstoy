package main
import (
	"gopkg.in/yaml.v3"
	"os"
	"log"
	"net"
)
func loadConfig(config_path string) () {
	//check if the file exists
	_ , err := os.Stat(config_path)
	if os.IsNotExist(err){
		log.Panicln("Couldnt find the config file")
	}

	config_file,_ := os.ReadFile(config_path)

	err = yaml.Unmarshal( config_file , &brokerSettings)

	if err != nil{
		log.Panicf("Failed to parse the error %v",err)
	}

	//add the predefine the topics
	for _,topicname := range brokerSettings.Topics{
		Topics[topicname] = make(map[net.Conn]struct{}) 
	}

	log.Println("Loaded config succesfully")
}
