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

var conf Config

func main() {
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

	t := NewTor(*conf.torPath)
	if err := t.Start(); err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := t.Stop(); err != nil {
			log.Fatal(err)
		}
	}()
}
