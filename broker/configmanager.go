package broker

import (
	"errors"
	"log"
	"net"
	"os"
	"fmt"

	"gopkg.in/yaml.v3"
)
func loadConfig(config_path string) (error) {
	//check if the file exists
	_ , err := os.Stat(config_path)
	if os.IsNotExist(err){
		return errors.New("could not find the config file")
	}

	config_file,_ := os.ReadFile(config_path)

	err = yaml.Unmarshal( config_file , &brokerSettings)

	if err != nil{
		return err
	}

	if brokerSettings.Port == -1{
		return errors.New("no port defined in the config file")
	}

	if brokerSettings.Host == ""{
		return errors.New("no host defined in the config file")
	}
	//add the predefine the topics
	for _,topicname := range brokerSettings.Topics{
		Topics[topicname] = make(map[net.Conn]bool) 
		indexFileName := brokerSettings.Persistence.Directory + topicname + ".index"
		lastOffset , err := getLastOffset(indexFileName)
		if err != nil {
			return fmt.Errorf("failed to read last offset for the file %s",indexFileName)
		}
		TopicOffsets[topicname] = lastOffset 
	}

	log.Println("Loaded config succesfully")
	return nil
}
