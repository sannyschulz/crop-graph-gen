package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sannyschulz/cropgraph"
)

func main() {
	fmt.Println("crop graph tool")
	inputFile := flag.String("input", "", "input file")
	configFile := flag.String("config", "", "config file")
	outputFile := flag.String("output", "", "output file")
	flag.Parse()

	// no config given
	if *configFile == "" {
		// write default config file
		cropgraph.WriteDefaultConfigFile("config.yml")
		return
		// if config file is given, but does not exist
	} else if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		fmt.Println("config file does not exist")
		// write default config file
		cropgraph.WriteDefaultConfigFile(*configFile)
		return
	}
	if *inputFile == "" {
		fmt.Println("no input file given")
		return
	}

	// 1. read the hermes simulation output file
	err := cropgraph.SefaultCsvToGraph(*inputFile, *configFile, *outputFile)
	if err != nil {
		fmt.Println(err)
	}

}
