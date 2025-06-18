package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Tervicke/Tolstoy/broker"
	"github.com/common-nighthawk/go-figure"
)
var VERSION = "dev"
var DATE = "unknown"
func main(){
	config_path := flag.String("config" , "config.yaml" ,"The path to the config.yaml file \n by default it searches the current directory")                              
	showVersion := flag.Bool("version" , false , "print the version and exit")
	flag.Parse()
	if(*showVersion){
		figure.NewFigure("Tolstoy", "", true).Print()
		fmt.Printf("Version: %s\n", VERSION)
		fmt.Printf("Build Date: %s\n", DATE)
		os.Exit(0)
	}
	broker.StartServer(*config_path)
}


