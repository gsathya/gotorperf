import (
	"crypto/rand"
	"encoding/base32"
	"net/http"
	"time"

	"golang.org/x/net/proxy"
)

func createSocksfiedHTTPClient(torAddr string, timeout time.Duration) (*http.Client, error) {
	// We generate a random username so that Tor will decouple all of our
	// connections.
	username := make([]byte, 8)
	if _, err := rand.Read(username); err != nil {
		return nil, err
	}

	auth := proxy.Auth{
		User:     base32.StdEncoding.EncodeToString(username),
		Password: "password",
	}

	dialer, err := proxy.SOCKS5("tcp", torAddr, &auth, proxy.Direct)
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{Dial: dialer.Dial}
	httpClient := &http.Client{Transport: tr, Timeout: iimeout}

	return httpClient, nil
}
