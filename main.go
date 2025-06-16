package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Tervicke/Tolstoy/broker"
)
const VERSION = "v0.4.0";

func main(){
	config_path := flag.String("config" , "config.yaml" ,"The path to the config.yaml file \n by default it searches the current directory")                              
	showVersion := flag.Bool("version" , false , "print the version and exit")
	flag.Parse()
	if(*showVersion){
		fmt.Println(VERSION)
		os.Exit(0)
	}
	broker.StartServer(*config_path)
}
