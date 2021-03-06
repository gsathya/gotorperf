package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"time"

	"github.com/gsathya/torperf/torctl"
)

const (
	httpBufLen = 16 // inherited from the original C codebase
	uri        = "https://torperf.torproject.org:80/%s"
	socksAddr  = "127.0.0.1:9050"
	ctrlAddr   = "127.0.0.1:9051"
	ctrlPort   = "9051"
	request    = "GET %s HTTP/1.0\r\nPragma: no-cache\r\n" +
		"Host: %s\r\n\r\n"
)

var (
	t       *torctl.Tor
	circs   = make(map[string]*torctl.CircEvent)
	streams = make(map[string]*torctl.StreamEvent)
)

type StaticFileDownload struct {
	Uri string

	Received int
	Expected int
	Sent     int

	Start        time.Time
	Datarequest  time.Duration
	Dataresponse time.Duration
	Datacomplete time.Duration
	Dataperctime [9]time.Duration
}

func (s *StaticFileDownload) run() (err error) {
	u, err := url.Parse(s.Uri)
	if err != nil {
		return err
	}

	s.Start = time.Now() //XXX: unix timestamp?
	log.Println("creating socksfied dialer")
	dialer, err := NewSocksfiedDialer(socksAddr)
	if err != nil {
		return err
	}

	log.Println("dialing tcp")
	conn, err := dialer.Dial("tcp", u.Host)
	if err != nil {
		return err
	}

	req := fmt.Sprintf(request, u.Path, u.Host)
	s.Sent = len(req)
	log.Printf("request: %s", req)

	log.Println("sending request")
	fmt.Fprintf(conn, req)
	// get when request is sent
	s.Datarequest = time.Since(s.Start)

	n := 0
	decile := -1
	buf := make([]byte, httpBufLen)

	log.Println("reading response")
	for {
		n, err = conn.Read(buf)

		// get when start of response was received
		if n > 0 && s.Received == 0 {
			s.Dataresponse = time.Since(s.Start)
		}

		s.Received += n

		// get when the next 10% of expected bytes are received; this is a
		// for loop for cases when we expect only very few bytes and read
		// more than 10% of them in a single read_all() call.
		for s.Received < s.Expected &&
			s.Received*10/s.Expected > decile+1 {
			decile += 1
			s.Dataperctime[decile] = time.Since(s.Start)
		}

		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
	}

	// get when response is complete
	s.Datacomplete = time.Since(s.Start)

	log.Println("total size of response", s.Received)
	log.Println("dataperctime", s.Dataperctime)
	return nil
}

func StaticFileExperimentRunner(c *Config) (result []byte, err error) {
	s := StaticFileDownload{
		Uri:      fmt.Sprintf(uri, ".50kbfile"),
		Expected: 51200,
	}

	torrc := make(map[string]string)
	torrc["controlport"] = ctrlPort

	// start tor
	t, err := torctl.StartWithConfig(*c.torPath, torrc)
	if err != nil {
		return nil, err
	}
	defer func() {
		if terr := t.Stop(); err != nil {
			err = terr
		}
	}()

	// connect to control port
	conn, err := torctl.Connect(ctrlAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to control port: %v", err)
	}
	defer func() {
		if cerr := conn.Close(); err != nil {
			err = cerr
		}
	}()

	// set up event handlers
	if err = conn.On(torctl.CIRC, handleCircEvent); err != nil {
		return nil, err
	}
	if err = conn.On(torctl.STREAM, handleStreamEvent); err != nil {
		return nil, err
	}

	// start experiment
	if err = s.run(); err != nil {
		return nil, err
	}

	log.Print("Finished experiment")
	return json.Marshal(s)
}

func handleStreamEvent(e torctl.Event) {
	ev := e.(*torctl.StreamEvent)

	log.Print("StreamEvent: ", ev)
	if ev.Status == "SETCONNECT" && ev.Target == "torperf.torproject.org:80" {
		streams[ev.Id] = ev
	}
	_, ok := streams[ev.Id]
	if !ok {
		return
	}
}

func handleCircEvent(e torctl.Event) {
	ev := e.(*torctl.CircEvent)

	log.Print("CircEvent: ", ev)
	if ev.Status == "LAUNCHED" {
		circs[ev.Id] = ev
	}
	_, ok := circs[ev.Id]
	if !ok {
		return
	}
}

func init() {
	experiments["static_file_download"] = StaticFileExperimentRunner
}
