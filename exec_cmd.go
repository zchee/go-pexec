// SPDX-FileCopyrightText: Copyright 2021 The go-pexec Authors
// SPDX-License-Identifier: BSD-3-Clause

package pexec

import (
	"context"
	"strings"

	exec "golang.org/x/sys/execabs"
)

type execCmd struct {
	*exec.Cmd
}

func newExecCmd(ctx context.Context, cmd *exec.Cmd) *execCmd {
	return &execCmd{cmd}
}

func (e *execCmd) Kill() error {
	if e.Process != nil {
		return e.Process.Kill()
	}
	return nil
}

func (e *execCmd) String() string {
	return strings.Join(append([]string{e.Path}, e.Args...), " ")
}
