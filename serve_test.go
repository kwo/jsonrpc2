// SPDX-FileCopyrightText: 2021 The Go Language Server Authors
// SPDX-FileCopyrightText: 2026 Karl Ostendorf
// SPDX-License-Identifier: BSD-3-Clause

package jsonrpc2_test

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/kwo/jsonrpc2"
)

func TestIdleTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := ln.Close(); err != nil {
			t.Error(err)
		}
	}()

	connect := func() net.Conn {
		d := net.Dialer{Timeout: 5 * time.Second}
		conn, err := d.DialContext(ctx, "tcp", ln.Addr().String())
		if err != nil {
			panic(err)
		}
		return conn
	}

	server := jsonrpc2.HandlerServer(jsonrpc2.MethodNotFoundHandler)
	var (
		runErr error
		wg     sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		runErr = jsonrpc2.Serve(ctx, ln, server, 100*time.Millisecond)
	}()

	// Exercise some connection/disconnection patterns, and then assert that when
	// our timer fires, the server exits.
	conn1 := connect()
	conn2 := connect()
	if err := conn1.Close(); err != nil {
		t.Error(err)
	}
	if err := conn2.Close(); err != nil {
		t.Error(err)
	}
	conn3 := connect()
	if err := conn3.Close(); err != nil {
		t.Error(err)
	}

	wg.Wait()

	if !errors.Is(runErr, jsonrpc2.ErrIdleTimeout) {
		t.Errorf("run() returned error %v, want %v", runErr, jsonrpc2.ErrIdleTimeout)
	}
}
