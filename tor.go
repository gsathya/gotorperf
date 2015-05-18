package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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
	buf := bufio.NewReader(stdout)

	if err := cmd.Start(); err != nil {
		return err
	}

	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return errors.New("tor did not bootstrap")
			}
			return err
		}

		if time.Now().After(timeout) {
			err = cmd.Process.Kill()
			if err != nil {
				return err
			}
			return errors.New("process killed because of timeout")
		}

		if strings.Contains(line, "Bootstrapped 100%: Done") {
			break
		}
	}
	return nil
}
