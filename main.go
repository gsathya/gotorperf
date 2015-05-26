package main

import (
	"flag"
	"log"
	"os"
)

type Config struct {
	logPath *string
	torPath *string
}

type Experiment func(c *Config) (err error)

var experiments map[string]Experiment

func init() {
	experiments = make(map[string]Experiment)
}

func main() {
	var conf Config
	// command line args
	conf.logPath = flag.String("log", "", "path to log file; otherwise stdout")
	conf.torPath = flag.String("tor", "", "path to tor binary; otherwise uses $PATH")
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
		if err := exp(&conf); err != nil {
			log.Print(err)
		}
	}
}
