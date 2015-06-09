package torctl

import (
	"fmt"
	"log"
	"strings"

	"github.com/Yawning/bulb"
)

type EventHandler func(e Event)

type Conn struct {
	c      *bulb.Conn
	muxer  map[string]EventHandler
	done   chan struct{}
	events []string
}

func Connect(ctrlAddr string) (*Conn, error) {
	var err error
	c := &Conn{
		muxer: make(map[string]EventHandler),
	}

	c.c, err = bulb.Dial("tcp4", ctrlAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to control port: %v", err)
	}

	// authenticate controlport
	if err := c.c.Authenticate(""); err != nil {
		return nil, fmt.Errorf("Authentication failed: %v", err)
	}

	c.c.StartAsyncReader()
	go c.handler()
	return c, nil
}

func (c *Conn) Close() error {
	//c.done <- struct{}{}
	return c.c.Close()
}

func (c *Conn) On(et string, eh EventHandler) error {
	c.events = append(c.events, et)

	cmd := fmt.Sprintf("SETEVENTS %s", strings.Join(c.events, " "))
	fmt.Println(cmd)
	if _, err := c.c.Request(cmd); err != nil {
		return err
	}

	c.muxer[et] = eh
	return nil
}

//XXX: leaky!
func (c *Conn) handler() {
	for {
		select {
		default:
			ev, err := c.c.NextEvent()

			if err != nil {
				log.Fatalf("NextEvent() failed: %v", err)
			}

			e, err := Parse(ev.Reply)
			if err != nil {
				log.Print(err)
				continue
			}

			t := e.Type()
			f, ok := c.muxer[t]
			if !ok {
				fmt.Print("Unknow type: %T", t)
				continue
			}

			f(e)
		case <-c.done:
			return
		}
	}

}
