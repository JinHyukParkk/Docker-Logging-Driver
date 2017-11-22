package main

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/docker/docker/daemon/logger"
	"github.com/docker/go-plugins-helpers/sdk"
)

// StartRequest is the format of requests coming from Docker to the
// /LogDriver.StartLogging endpoint.
type StartRequest struct {
	File string      // FIFO file path set up by Docker. Log messages will be written to it.
	Info logger.Info // The struct defined by Docker. Can only depend on ContainerID being set.
}

// StopRequest is the format of requests coming from Docker to the
// /LogDriver.StopLogging endpoint.
type StopRequest struct {
	File string // Corresponds to a FIFO path sent in a StartRequest by Docker.
}

// CapabilitiesResponse is the format of requests coming from Docker to the
// /LogDriver.Capabilities endpoint
type CapabilitiesResponse struct {
	Err string
	Cap logger.Capability // The struct defined by Docker. Only the ReadLogs field is supported for now.
}

// ErrorResponse is the format of responses to Docker when an error occurs.
type ErrorResponse struct {
	Err string
}

func respond(err error, w http.ResponseWriter) {
	var r ErrorResponse
	if err != nil {
		r.Err = err.Error()
	}
	json.NewEncoder(w).Encode(&r)
}

func inithandlers(h *sdk.Handler, d LoggingDriver) {
	h.HandleFunc("/LogDriver.StartLogging", func(w http.ResponseWriter, r *http.Request) {
		var (
			err      error
			startreq StartRequest
		)
		if err = json.NewDecoder(r.Body).Decode(&startreq); err != nil {
			err = errors.Wrap(err, "error unmarshalling request body in /LogDriver.StartLogging handler")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if startreq.Info.ContainerID == "" {
			respond(errors.New("ContainerID field is required in requests to /LogDriver.StartLogging"), w)
			return
		}
		err = d.StartLogging(startreq.File, startreq.Info)
		respond(err, w)
	})

	h.HandleFunc("/LogDriver.StopLogging", func(w http.ResponseWriter, r *http.Request) {
		var (
			err     error
			stopreq StopRequest
		)
		if err = json.NewDecoder(r.Body).Decode(&stopreq); err != nil {
			err = errors.Wrap(err, "error unmarshalling request body in /LogDriver.StopLogging handler")
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		err = d.StopLogging(stopreq.File)
		respond(err, w)
	})

	h.HandleFunc("/LogDriver.Capabilities", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(&CapabilitiesResponse{
			Cap: logger.Capability{ReadLogs: false},
		})
	})
}
