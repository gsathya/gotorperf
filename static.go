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
	uri           = "https://torperf.torproject.org:80/.50kbfile"
	torAddr       = "127.0.0.1:9050"
	request       = "GET %s HTTP/1.0\r\nPragma: no-cache\r\n" +
		"Host: %s\r\n\r\n"
)

var startTime time.Time

type StaticFileExperiment struct{}

func (s StaticFileExperiment) Run() (err error) {
	sfd := StaticFileDownload{}
	if err = sfd.run(); err != nil {
		return err
	}
	return
}

type StaticFileDownload struct {
	receivedBytes int
	sentBytes     int

	payloadSize  int
	expectedSize int

	datarequest  time.Duration
	dataresponse time.Duration
	datacomplete time.Duration
	dataperctime []time.Duration
}

func (s *StaticFileDownload) ReadFrom(r io.Reader) (err error) {
	p := make([]byte, http_read_len)
	s.dataperctime = make([]time.Duration, 9)
	s.receivedBytes = 0
	s.expectedSize = 51200
	decile := -1

	log.Println("reading response")
	for {
		n, err := r.Read(p)
		s.receivedBytes += n
		for s.receivedBytes < s.expectedSize &&
			s.receivedBytes*10/s.expectedSize > decile+1 {
			decile += 1
			s.dataperctime[decile] = time.Since(startTime)
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
	t := torctl.NewTor(*conf.torPath)
	if err = t.Start(); err != nil {
		return err
	}
	defer func() {
		if terr := t.Stop(); err != nil {
			err = terr
		}
	}()

	u, err := url.Parse(uri)
	if err != nil {
		return err
	}

	startTime = time.Now()
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
	log.Printf("request: %s", req)

	log.Println("sending request")
	fmt.Fprintf(conn, req)
	s.datarequest = time.Since(startTime)

	err = s.ReadFrom(conn)
	if err != nil {
		return err
	}
	s.datacomplete = time.Since(startTime)
	log.Println("total size of response", s.receivedBytes)
	log.Println("dataperctime", s.dataperctime)
	return
}

func init() {
	experiments["static_file_download"] = StaticFileExperiment{}
}
