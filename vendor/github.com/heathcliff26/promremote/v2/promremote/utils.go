package promremote

import (
	"net/http"
	"os"
)

type basicAuthRoundTripper struct {
	username string
	password string
}

func (rt *basicAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(rt.username, rt.password)
	return http.DefaultTransport.RoundTrip(req)
}

func addBasicAuthToHTTPClient(c *http.Client, username, password string) {
	c.Transport = &basicAuthRoundTripper{
		username: username,
		password: password,
	}
}

func newHTTPClientWithTimeout() *http.Client {
	return &http.Client{
		Timeout: httpClientTimeout,
	}
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err == nil {
		return hostname
	}
	return "localhost"
}
