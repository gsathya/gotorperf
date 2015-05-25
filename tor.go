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

type Tor struct {
	path    string
	timeout time.Duration
	cmd     *exec.Cmd
	running bool
}

func NewTor(torPath string) *Tor {
	return &Tor{
		torPath,
		90,
		nil,
		false,
	}
}

func (t *Tor) Start() error {
	var err error
	log.Println("starting tor")

	if t.path == "" {
		t.path, err = exec.LookPath("tor")
		if err != nil {
			return errors.New("tor not found in $PATH")
		}
	}

	f, err := os.Stat(t.path)
	if err != nil {
		return err
	}

	if f.IsDir() {
		return fmt.Errorf(t.path, " is a directory, not the tor executable")
	}

	t.cmd = exec.Command(t.path)
	stdout, err := t.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	s := bufio.NewScanner(stdout)

	timeout := time.Now().Add(init_timeout)
	if err := t.cmd.Start(); err != nil {
		return err
	}

	for s.Scan() {
		line := s.Text()

		if strings.Contains(line, "Bootstrapped 100%: Done") {
			t.running = true
			return nil
		}

		if time.Now().After(timeout) {
			err = t.cmd.Process.Kill()
			if err != nil {
				return err
			}
			return errors.New("tor process killed because of timeout")
		}
	}

	return errors.New("tor did not bootstrap")
}

func (t *Tor) Stop() error {
	if err := t.cmd.Process.Kill(); err != nil {
		return err
	}
	return nil
}
