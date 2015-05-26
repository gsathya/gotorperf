package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"time"

	"github.com/gsathya/torperf/torctl"
)

const (
	uri     = "https://torperf.torproject.org:80/.50kbfile"
	torAddr = "127.0.0.1:9050"
	request = "GET %s HTTP/1.0\r\nPragma: no-cache\r\n" +
		"Host: %s\r\n\r\n"
)

type StaticFileExperiment struct{}

func (s StaticFileExperiment) Run() (err error) {
	sfd := StaticFileDownload{}
	if err = sfd.run(); err != nil {
		return err
	}
	return
}

type StaticFileDownload struct {
	datarequest  time.Duration
	dataresponse time.Duration
	datacomplete time.Duration
}

func (s StaticFileDownload) run() (err error) {
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

	log.Println("reading response")
	n, err := io.Copy(ioutil.Discard, conn)
	if err != nil {
		return err
	}
	log.Println("total size of response", n)
	return
}

func init() {
	experiments["static_file_download"] = StaticFileExperiment{}
}
