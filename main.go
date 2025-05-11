package main

import (
	"Tolstoy/broker"
	"flag"
)
func main(){
	config_path := flag.String("config" , "config.yaml" ,"The path to the config.yaml file \n by default it searches the current directory")                              
	flag.Parse()
	broker.StartServer(*config_path)
}
