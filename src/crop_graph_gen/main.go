package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sannyschulz/crop-graph-gen/cropgraph"
)

func main() {
	fmt.Println("crop graph tool")
	// read command line arguments
	inputFile := flag.String("input", "", "input file")
	batchFile := flag.String("batch", "", "batch file")
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

	if *inputFile == "" && *batchFile == "" {
		fmt.Println("no input file given")
		return
	}
	if *batchFile != "" {
		// a batch file contains a list of input and output files as comma separated csv file
		// read the batch file and process each line
		err := cropgraph.BatchFileToGraph(*batchFile, *configFile)
		if err != nil {
			fmt.Println(err)
		}
	} else {

		// read config
		config, err := cropgraph.ReadConfigFile(*configFile)
		if err != nil {
			fmt.Println(err)
			return
		}

		// read the hermes simulation output file
		err = cropgraph.HermesCsvToGraph(*inputFile, config, *outputFile)
		if err != nil {
			fmt.Println(err)
		}
	}
}
