// SPDX-FileCopyrightText: 2026 Karl Ostendorf
// SPDX-License-Identifier: BSD-3-Clause

package jsonrpc2_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/kwo/jsonrpc2"
)

// TestMain re-execs the test binary as a JSON-RPC server when the
// JSONRPC2_SERVER env var is set. Otherwise it runs tests normally.
func TestMain(m *testing.M) {
	if os.Getenv("JSONRPC2_SERVER") != "" {
		serveStdio()
		return
	}
	os.Exit(m.Run())
}

// serveStdio runs a JSON-RPC server on stdin/stdout.
func serveStdio() {
	framer := jsonrpc2.HeaderFramer()
	conn := jsonrpc2.NewConn(framer.Reader(os.Stdin), framer.Writer(os.Stdout), os.Stdin)
	conn.Go(context.Background(), echoHandler)
	<-conn.Done()
}

// echoHandler handles "echo" calls by returning the params and "ping" calls
// by returning "pong". Notifications are silently accepted.
func echoHandler(_ context.Context, req jsonrpc2.Request) (any, error) {
	switch req.Method() {
	case "echo":
		var v json.RawMessage
		if err := json.Unmarshal(req.Params(), &v); err != nil {
			return nil, err
		}
		return v, nil
	case "ping":
		return "pong", nil
	default:
		return jsonrpc2.MethodNotFoundHandler(context.Background(), req)
	}
}

// startServer launches a child process running the JSON-RPC server and
// returns a Conn connected to it.
func startServer(t *testing.T) jsonrpc2.Conn {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = append(os.Environ(), "JSONRPC2_SERVER=1")
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	framer := jsonrpc2.HeaderFramer()
	conn := jsonrpc2.NewConn(framer.Reader(stdout), framer.Writer(stdin), stdin)
	conn.Go(context.Background(), jsonrpc2.MethodNotFoundHandler)

	t.Cleanup(func() {
		conn.Close()
		<-conn.Done()
		_ = cmd.Wait()
	})

	return conn
}

func TestStdioCall(t *testing.T) {
	conn := startServer(t)
	ctx := context.Background()

	var got string
	if _, err := conn.Call(ctx, "ping", nil, &got); err != nil {
		t.Fatal(err)
	}
	if got != "pong" {
		t.Fatalf("got %q, want %q", got, "pong")
	}
}

func TestStdioAsyncCall(t *testing.T) {
	conn := startServer(t)
	ctx := context.Background()

	ac1, err := conn.AsyncCall(ctx, "echo", "hello")
	if err != nil {
		t.Fatal(err)
	}
	ac2, err := conn.AsyncCall(ctx, "echo", "world")
	if err != nil {
		t.Fatal(err)
	}

	var r1, r2 string
	if err := ac1.Await(ctx, &r1); err != nil {
		t.Fatal(err)
	}
	if err := ac2.Await(ctx, &r2); err != nil {
		t.Fatal(err)
	}

	if r1 != "hello" || r2 != "world" {
		t.Fatalf("got %q %q, want %q %q", r1, r2, "hello", "world")
	}
}

func TestStdioNotify(t *testing.T) {
	conn := startServer(t)
	ctx := context.Background()

	// Notify doesn't expect a response; just verify it doesn't error.
	if err := conn.Notify(ctx, "ping", nil); err != nil {
		t.Fatal(err)
	}

	// Verify the connection still works after a notification.
	var got string
	if _, err := conn.Call(ctx, "ping", nil, &got); err != nil {
		t.Fatal(err)
	}
	if got != "pong" {
		t.Fatalf("got %q, want %q", got, "pong")
	}
}

func ExampleConn_stdio() {
	// Start a child process running the JSON-RPC server.
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = append(os.Environ(), "JSONRPC2_SERVER=1")

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	_ = cmd.Start()

	// Connect using HeaderFramer over the child's stdin/stdout.
	framer := jsonrpc2.HeaderFramer()
	conn := jsonrpc2.NewConn(framer.Reader(stdout), framer.Writer(stdin), stdin)
	conn.Go(context.Background(), jsonrpc2.MethodNotFoundHandler)

	// Synchronous call.
	var pong string
	conn.Call(context.Background(), "ping", nil, &pong)
	fmt.Println(pong)

	// Async calls.
	ac1, _ := conn.AsyncCall(context.Background(), "echo", "hello")
	ac2, _ := conn.AsyncCall(context.Background(), "echo", "world")
	var r1, r2 string
	ac1.Await(context.Background(), &r1)
	ac2.Await(context.Background(), &r2)
	fmt.Println(r1, r2)

	conn.Close()
	<-conn.Done()
	cmd.Wait()

	// Output:
	// pong
	// hello world
}
