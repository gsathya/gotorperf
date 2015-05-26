package main

import (
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
	start time.Time
	t     *torctl.Tor
)

func StaticFileExperimentRunner(c *Config) (err error) {
	s := StaticFileDownload{
		uri:          fmt.Sprintf(uri, ".50kbfile"),
		expected:     51200,
		dataperctime: make([]time.Duration, 9),
	}

	t := torctl.NewTor(*c.torPath)
	if err = t.Start(); err != nil {
		return err
	}
	defer func() {
		if terr := t.Stop(); err != nil {
			err = terr
		}
	}()

	if err = s.run(); err != nil {
		return
	}
	return
}

type StaticFileDownload struct {
	uri string

	received int
	expected int
	sent     int

	datarequest  time.Duration
	dataresponse time.Duration
	datacomplete time.Duration
	dataperctime []time.Duration
}

func (s *StaticFileDownload) ReadFrom(r io.Reader) (err error) {
	var (
		buf    = make([]byte, http_read_len)
		decile = -1
	)

	log.Println("reading response")
	for {
		n, err := r.Read(buf)

		// Get when start of response was received
		if n > 0 && s.received == 0 {
			s.dataresponse = time.Since(start)
		}

		s.received += n

		// Get when the next 10% of expected bytes are received; this is a
		// while loop for cases when we expect only very few bytes and read
		// more than 10% of them in a single read_all() call.
		for s.received < s.expected &&
			s.received*10/s.expected > decile+1 {
			decile += 1
			s.dataperctime[decile] = time.Since(start)
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

func (s *StaticFileDownload) run() (err error) {
	u, err := url.Parse(s.uri)
	if err != nil {
		return err
	}

	start = time.Now()
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
	s.sent = len(req)
	log.Printf("request: %s", req)

	log.Println("sending request")
	fmt.Fprintf(conn, req)
	// Get when request is sent
	s.datarequest = time.Since(start)

	if err = s.ReadFrom(conn); err != nil {
		return err
	}
	// Get when response is complete
	s.datacomplete = time.Since(start)

	log.Println("total size of response", s.received)
	log.Println("dataperctime", s.dataperctime)
	return
}

func init() {
	experiments["static_file_download"] = StaticFileExperimentRunner
}
