// SPDX-FileCopyrightText: Copyright 2021 The go-pexec Authors
// SPDX-License-Identifier: BSD-3-Clause

package pexec

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"time"
)

var errInterrupted = errors.New("runner interrupted by signal")

type runner struct {
	FastFail          bool
	MaxConcurrentCmds int
	EventHandler      func(*Event)
	Clock             func() time.Time
}

func newRunner(options ...RunnerOption) *runner {
	runner := &runner{
		DefaultFastFail,
		DefaultMaxConcurrentCmds,
		DefaultEventHandler,
		DefaultClock,
	}
	for _, option := range options {
		option(runner)
	}
	return runner
}

func (r *runner) Run(cmds []Cmd) error {
	// there is a race condition where err could be set to
	// errCmdFailed or not set at all even after an interrupt happens
	var err error
	doneC := make(chan struct{})
	cmdControllers := make([]*cmdController, len(cmds))
	for i, cmd := range cmds {
		cmdControllers[i] = newCmdController(cmd, r.EventHandler, r.Clock)
	}

	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, os.Interrupt)
	go func() {
		for range signalC {
			// do not want to acquire lock in the signal handler
			err = errInterrupted
			doneC <- struct{}{}
			return
		}
	}()

	var wg sync.WaitGroup
	semaphore := newSemaphore(r.MaxConcurrentCmds)

	startTime := r.Clock()
	r.EventHandler(newStartedEvent(startTime))
	for _, cmdController := range cmdControllers {
		cmdController := cmdController
		wg.Add(1)
		go func() {
			semaphore.P(1)
			defer semaphore.V(1)
			defer wg.Done()
			if !cmdController.Run() {
				// best effort to prioritize the interrupt error
				// but this is not deterministic
				err = errCmdFailed
				if r.FastFail {
					doneC <- struct{}{}
				}
			}
		}()
	}
	go func() {
		// if everything finishes and there is an interrupt, we could
		// end up not actually returning an error if everything below
		// completes before we context switch to the interrupt goroutine
		wg.Wait()
		doneC <- struct{}{}
	}()
	// this waits on command completion, fast failure, or signal
	<-doneC
	for _, cmdController := range cmdControllers {
		cmdController.Kill()
	}
	finishTime := r.Clock()
	r.EventHandler(newFinishedEvent(finishTime, startTime, err))
	return err
}
