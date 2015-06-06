package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

type Config struct {
	logPath     *string
	torPath     *string
	outputPath  *string
	experiments *string
}

type Experiment func(c *Config) (result []byte, err error)

var experiments map[string]Experiment

func init() {
	experiments = make(map[string]Experiment)
}

func main() {
	var (
		conf             Config
		experimentsToRun []string
	)

	// command line args
	conf.logPath = flag.String("log", "", "path to log file; otherwise stdout")
	conf.torPath = flag.String("tor", "", "path to tor binary; otherwise uses $PATH")
	conf.outputPath = flag.String("output", "", "path to output file; otherwise uses stdout")
	conf.experiments = flag.String("experiments", "", "list of experiments to run; otherwise runs all")
	flag.Parse()

	// log to file
	if len(*conf.logPath) > 0 {
		f, err := os.Create(*conf.logPath)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(f)
	}

	// run only some experiments
	if len(*conf.experiments) > 0 {
		experimentsToRun = strings.Split(*conf.experiments, ",")
	} else {
		// run all experiments
		for name := range experiments {
			experimentsToRun = append(experimentsToRun, name)
		}
	}

	for _, name := range experimentsToRun {
		exp, ok := experiments[name]
		if !ok {
			log.Printf("experiment: %s not found", name)
			continue
		}

		log.Printf("running experiment: %s", name)
		result, err := exp(&conf)
		if err != nil {
			log.Print(err)
			continue
		}
		log.Println(string(result))
	}
}
