package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var init_timeout = 90 * time.Second

func startTor(torPath string) error {
	var err error
	log.Println("starting tor")

	if torPath == "" {
		torPath, err = exec.LookPath("tor")
		if err != nil {
			return errors.New("tor not found in $PATH")
		}
	}

	f, err := os.Stat(torPath)
	if err != nil {
		return err
	}

	if f.IsDir() {
		return fmt.Errorf(torPath, " is a directory, not the tor executable")
	}

	timeout := time.Now().Add(init_timeout)
	cmd := exec.Command(torPath)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	s := bufio.NewScanner(stdout)

	if err := cmd.Start(); err != nil {
		return err
	}

	for s.Scan() {
		line := s.Text()
		if strings.Contains(line, "Bootstrapped 100%: Done") {
			return nil
		}

		if time.Now().After(timeout) {
			err = cmd.Process.Kill()
			if err != nil {
				return err
			}
			return errors.New("tor process killed because of timeout")
		}
	}

	return errors.New("tor did not bootstrap")
}
