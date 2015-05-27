package main

import (
	"flag"
	"log"
	"os"
)

type Config struct {
	logPath    *string
	torPath    *string
	outputPath *string
}

type Experiment func(c *Config) (result []byte, err error)

var experiments map[string]Experiment

func init() {
	experiments = make(map[string]Experiment)
}

func main() {
	var conf Config

	// command line args
	conf.logPath = flag.String("log", "", "path to log file; otherwise stdout")
	conf.torPath = flag.String("tor", "", "path to tor binary; otherwise uses $PATH")
	conf.outputPath = flag.String("output", "", "path to output file; otherwise uses stdout")
	flag.Parse()

	// log to file
	if len(*conf.logPath) > 0 {
		f, err := os.Create(*conf.logPath)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(f)
	}

	for name, exp := range experiments {
		log.Printf("running experiment: %s", name)
		result, err := exp(&conf)
		if err != nil {
			log.Print(err)
			continue
		}

		log.Println(string(result))
	}
}
