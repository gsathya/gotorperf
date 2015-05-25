package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gsathya/torperf/torctl"
)

var url = "http://torperf.torproject.org/.50kbfile"
var files = []string{".50kbfile"}

type StaticFileDownload struct{}

func (s StaticFileDownload) Run() (err error) {
	t := torctl.NewTor(*conf.torPath)
	if err = t.Start(); err != nil {
		return err
	}
	defer func() {
		if terr := t.Stop(); err != nil {
			err = terr
		}
	}()

	client, err := NewSocksfiedHTTPClient("127.0.0.1:9050", 0)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/octet-stream")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	n, err := io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return err
	}
	log.Println("total size of payload", n)
	return
}

func init() {
	experiments["static_file_download"] = StaticFileDownload{}
}
