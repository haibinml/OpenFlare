// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
	"time"
)

func TestGoRunsFn(t *testing.T) {
	done := make(chan struct{})
	Go(func() { close(done) })
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("fn was not run")
	}
}

func TestGoRecoversPanic(t *testing.T) {
	done := make(chan struct{})
	Go(func() {
		defer close(done)
		panic("boom")
	})
	<-done
	// Give the recovering goroutine a moment to finish logging; the test
	// only fails if the panic had propagated and crashed the process.
	time.Sleep(10 * time.Millisecond)
}
