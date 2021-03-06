package torctl

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const timeout = 90 * time.Second

//XXX: Support multi-value config
type Torrc map[string]string

type Tor struct {
	path    string
	args    []string
	cmd     *exec.Cmd
	running bool
}

func StartWithConfig(torPath string, c Torrc) (*Tor, error) {
	t := &Tor{
		path: torPath,
	}

	f, err := ioutil.TempFile("", "torrc")
	if err != nil {
		return nil, err
	}
	defer func() {
		if ferr := f.Close(); err != nil {
			err = ferr
		}

		if ferr := os.Remove(f.Name()); err != nil {
			err = ferr
		}
	}()

	// we need to be sure that we're logging to stdout to figure out when we're
	// done bootstrapping
	fmt.Fprintf(f, "Log NOTICE stdout\n")

	for key, val := range c {
		fmt.Fprintf(f, "%s %s\n", key, val)
	}

	t.args = append(t.args, "-f", f.Name())
	return t, t.start()
}

func Start(torPath string) (*Tor, error) {
	t := &Tor{
		path: torPath,
	}

	return t, t.start()
}

func (t *Tor) start() error {
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

	t.cmd = exec.Command(t.path, t.args...)
	stdout, err := t.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	s := bufio.NewScanner(stdout)

	timeout := time.Now().Add(timeout)
	if err := t.cmd.Start(); err != nil {
		return err
	}

	for s.Scan() {
		line := s.Text()
		log.Println(line)
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
	log.Println("stopping tor")
	if !t.running {
		return nil
	}
	if err := t.cmd.Process.Kill(); err != nil {
		return err
	}
	t.running = false
	return nil
}
