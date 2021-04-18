// SPDX-FileCopyrightText: Copyright 2021 The go-pexec Authors
// SPDX-License-Identifier: BSD-3-Clause

package pexec

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	json "github.com/goccy/go-json"
	exec "golang.org/x/sys/execabs"
)

// DefaultFastFail is the default value for fast fail.
const DefaultFastFail = false

var (
	// DefaultMaxConcurrentCmds is the default value for the maximum
	// number of concurrent commands.
	DefaultMaxConcurrentCmds = runtime.NumCPU()
	// DefaultEventHandler is the default Event handler.
	DefaultEventHandler = logEvent
	// DefaultClock is the default function to use as a clock.
	DefaultClock = time.Now
)

// Event is an event that happens during the runner's Run call.
type Event struct {
	Type   EventType              `json:"type,omitempty" yaml:"type,omitempty"`
	Time   time.Time              `json:"time,omitempty" yaml:"time,omitempty"`
	Fields map[string]interface{} `json:"fields,omitempty" yaml:"fields,omitempty"`
	Error  string                 `json:"error,omitempty" yaml:"error,omitempty"`
}

// RunnerOption is an option for a new Runner.
type RunnerOption func(*runner)

// WithFastFail returns a RunnerOption that will return error fun
// Run as soon as one of the commands fails.
func WithFastFail() RunnerOption {
	return func(runner *runner) {
		runner.FastFail = true
	}
}

// WithMaxConcurrentCmds returns a RunnerOption that will make the
// Runner only run maxConcurrentCmds at once, or unlimited if 0.
func WithMaxConcurrentCmds(maxConcurrentCmds int) RunnerOption {
	return func(runner *runner) {
		runner.MaxConcurrentCmds = maxConcurrentCmds
	}
}

// WithEventHandler returns a RunnerOption that will use the
// given EventHandler.
func WithEventHandler(eventHandler func(*Event)) RunnerOption {
	return func(runner *runner) {
		runner.EventHandler = eventHandler
	}
}

// WithClock returns a RunnerOption that will make the Runner
// use the given Clock.
func WithClock(clock func() time.Time) RunnerOption {
	return func(runner *runner) {
		runner.Clock = clock
	}
}

// Cmd is a command to run.
type Cmd interface {
	fmt.Stringer

	// Start the command.
	Start() error
	// Wait for the command to complete and block.
	Wait() error
	// Kill the command.
	Kill() error
}

// ExecCmd returns a new Cmd for the given exec.Cmd.
func ExecCmd(ctx context.Context, cmd *exec.Cmd) Cmd {
	return newExecCmd(ctx, cmd)
}

// ExecCmds returns a slice of Cmds for the given exec.Cmds.
func ExecCmds(ctx context.Context, cmds []*exec.Cmd) []Cmd {
	execCmds := make([]Cmd, len(cmds))
	for i, cmd := range cmds {
		execCmds[i] = ExecCmd(ctx, cmd)
	}
	return execCmds
}

// Runner runs the commands.
type Runner interface {
	// Run the commands.
	//
	// Return error if there was an initialization error, or any of
	// the running commands returned with a non-zero exit code.
	Run(cmds []Cmd) error
}

// NewRunner returns a new Runner.
func NewRunner(options ...RunnerOption) Runner {
	return newRunner(options...)
}

func logEvent(event *Event) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Print(event.Type, " ", err)
		return
	}
	log.Print(string(data))
}
