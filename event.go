// SPDX-FileCopyrightText: Copyright 2021 The go-pexec Authors
// SPDX-License-Identifier: BSD-3-Clause

package pexec

import "time"

func newEvent(e EventType, t time.Time, f map[string]interface{}, err error) *Event {
	var errString string
	if err != nil {
		errString = err.Error()
	}
	return &Event{e, t, f, errString}
}

func newStartedEvent(t time.Time) *Event {
	return newEvent(EventTypeStarted, t, nil, nil)
}

func newCmdStartedEvent(t time.Time, cmd Cmd) *Event {
	return newEvent(EventTypeCmdStarted, t, map[string]interface{}{
		"cmd": cmd.String(),
	}, nil)
}

func newCmdFinishedEvent(t time.Time, cmd Cmd, startTime time.Time, err error) *Event {
	return newEvent(EventTypeCmdFinished, t, map[string]interface{}{
		"cmd":      cmd.String(),
		"duration": t.Sub(startTime).String(),
	}, err)
}

func newFinishedEvent(t time.Time, startTime time.Time, err error) *Event {
	return newEvent(EventTypeFinished, t, map[string]interface{}{
		"duration": t.Sub(startTime).String(),
	}, err)
}
