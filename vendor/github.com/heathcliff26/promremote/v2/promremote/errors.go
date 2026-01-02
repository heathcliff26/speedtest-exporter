package promremote

import (
	"fmt"
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

type ErrFailedToCreateRemoteAPI struct {
	err error
}

func NewErrFailedToCreateRemoteAPI(err error) error {
	return &ErrFailedToCreateRemoteAPI{err: err}
}

func (e *ErrFailedToCreateRemoteAPI) Error() string {
	return fmt.Sprintf("Failed to create remote write API: %v", e.err)
}

type ErrClientAlreadyRunning struct{}

func (e ErrClientAlreadyRunning) Error() string {
	return "Only a single instance of the client can run at a time"
}
