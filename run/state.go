package run

import (
	"fmt"
	"time"

	"github.com/taskcluster/taskcluster-worker-runner/cfg"
	"github.com/taskcluster/taskcluster-worker-runner/files"
	taskcluster "github.com/taskcluster/taskcluster/clients/client-go/v15"
)

// State represents the state of the worker run.  Its contents are built up
// bit-by-bit during the start-worker process.
type State struct {
	// Information about the Taskcluster deployment where this
	// worker is runing
	RootURL string

	// Credentials for the worker, and their expiration time.  Shortly before
	// this expiration, worker-runner will try to gracefully stop the worker
	Credentials       taskcluster.Credentials
	CredentialsExpire time.Time

	// Information about this worker
	WorkerPoolID string
	WorkerGroup  string
	WorkerID     string

	// metadata from the provider (useful to display to the user for
	// debugging).  This is also a public API for docker-worker,
	// which doesn't, strictly speaking, require these fields,
	// but may fall onto undesirable defaults if these are not provided
	// A bit more info on that here
	// https://github.com/taskcluster/taskcluster-worker-runner/pull/30#pullrequestreview-277378260
	ProviderMetadata map[string]string

	// the accumulated WorkerConfig for this run, including files to create
	WorkerConfig *cfg.WorkerConfig
	Files        []files.File
}

// Check that the provided provided the information it was supposed to.
func (state *State) CheckProviderResults() error {
	if state.RootURL == "" {
		return fmt.Errorf("provider did not set RootURL")
	}

	if state.Credentials.ClientID == "" {
		return fmt.Errorf("provider did not set Credentials.ClientID")
	}

	if state.WorkerPoolID == "" {
		return fmt.Errorf("provider did not set WorkerPoolID")
	}

	if state.WorkerGroup == "" {
		return fmt.Errorf("provider did not set WorkerGroup")
	}

	if state.WorkerID == "" {
		return fmt.Errorf("provider did not set WorkerID")
	}

	return nil
}