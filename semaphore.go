// SPDX-FileCopyrightText: Copyright 2021 The go-pexec Authors
// SPDX-License-Identifier: BSD-3-Clause

package pexec

type semaphore chan struct{}

func newSemaphore(n int) semaphore {
	if n <= 0 {
		return nil
	}
	s := make(semaphore, n)
	for i := 0; i < n; i++ {
		s <- struct{}{}
	}
	return s
}

func (s semaphore) P(n int) {
	if s == nil {
		return
	}
	for i := 0; i < n; i++ {
		<-s
	}
}

func (s semaphore) V(n int) {
	if s == nil {
		return
	}
	for i := 0; i < n; i++ {
		s <- struct{}{}
	}
}
