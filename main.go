package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	// command line args
	logPath := flag.String("log", "", "path to log file; otherwise stdout")
	torPath := flag.String("tor", "", "path to tor binary; otherwise use $PATH")
	flag.Parse()

	// log to file
	if len(*logPath) > 0 {
		f, err := os.Create(*logPath)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(f)
	}

	t := NewTor(*torPath)
	if err := t.Start(); err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := t.Stop(); err != nil {
			log.Fatal(err)
		}
	}()
}
