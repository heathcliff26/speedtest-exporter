package promremote

import (
	"fmt"
	"io"
)

type ErrMissingEndpoint struct{}

func (e ErrMissingEndpoint) Error() string {
	return "No endpoint for prometheus remote_write provided"
}

type ErrMissingInstance struct{}

func (e ErrMissingInstance) Error() string {
	return "No instance name provided"
}

type ErrMissingJob struct{}

func (e ErrMissingJob) Error() string {
	return "No job name provided"
}

type ErrMissingRegistry struct{}

func (e ErrMissingRegistry) Error() string {
	return "No prometheus registry provided"
}

type ErrRemoteWriteFailed struct {
	StatusCode int
	Body       string
}

func NewErrRemoteWriteFailed(status int, resBody io.ReadCloser) *ErrRemoteWriteFailed {
	var body string
	b, err := io.ReadAll(resBody)
	if err != nil {
		body = err.Error()
	} else {
		body = string(b)
	}
	return &ErrRemoteWriteFailed{
		StatusCode: status,
		Body:       body,
	}
}

func (e *ErrRemoteWriteFailed) Error() string {
	return fmt.Sprintf("Prometheus remote_write returned with Status Code %d, expected 200. Response body: %s", e.StatusCode, e.Body)
}

type ErrMissingAuthCredentials struct{}

func (e ErrMissingAuthCredentials) Error() string {
	return "Need both username and password, at least one of them is empty"
}

type ErrInvalidMetricDesc struct {
	Desc string
}

func (e *ErrInvalidMetricDesc) Error() string {
	return "Received metric with invalid description: " + e.Desc
}
