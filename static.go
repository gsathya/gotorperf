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
	http_read_len = 16
	uri           = "https://torperf.torproject.org:80/%s"
	torAddr       = "127.0.0.1:9050"
	request       = "GET %s HTTP/1.0\r\nPragma: no-cache\r\n" +
		"Host: %s\r\n\r\n"
)

var (
	t *torctl.Tor
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
	Dataperctime []time.Duration
}

func (s *StaticFileDownload) run() (err error) {
	u, err := url.Parse(s.Uri)
	if err != nil {
		return err
	}

	s.Start = time.Now()
	log.Println("creating socksfied dialer")
	dialer, err := NewSocksfiedDialer(torAddr)
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
	// Get when request is sent
	s.Datarequest = time.Since(s.Start)

	if err = s.read(conn); err != nil {
		return err
	}
	// Get when response is complete
	s.Datacomplete = time.Since(s.Start)

	log.Println("total size of response", s.Received)
	log.Println("dataperctime", s.Dataperctime)
	return
}

func (s *StaticFileDownload) read(r io.Reader) (err error) {
	var (
		n      = 0
		buf    = make([]byte, http_read_len)
		decile = -1
	)

	log.Println("reading response")
	for {
		n, err = r.Read(buf)

		// Get when start of response was received
		if n > 0 && s.Received == 0 {
			s.Dataresponse = time.Since(s.Start)
		}

		s.Received += n

		// Get when the next 10% of expected bytes are received; this is a
		// while loop for cases when we expect only very few bytes and read
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
	return nil
}

func StaticFileExperimentRunner(c *Config) (result []byte, err error) {
	s := StaticFileDownload{
		Uri:          fmt.Sprintf(uri, ".50kbfile"),
		Expected:     51200,
		Dataperctime: make([]time.Duration, 9),
	}

	t := torctl.NewTor(*c.torPath)
	if err = t.Start(); err != nil {
		return nil, err
	}
	defer func() {
		if terr := t.Stop(); err != nil {
			err = terr
		}
	}()

	if err = s.run(); err != nil {
		return nil, err
	}

	return json.Marshal(s)
}

func init() {
	experiments["static_file_download"] = StaticFileExperimentRunner
}
