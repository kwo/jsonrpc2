// Command plugin is a simple JSON-RPC 2.0 server that reads from stdin
// and writes to stdout. It handles two methods:
//
//   - "greet"  — expects a string name, returns "Hello, <name>!"
//   - "add"    — expects [a, b] numbers, returns their sum
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kwo/jsonrpc2"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	framer := jsonrpc2.RawFramer()
	conn := jsonrpc2.NewConn(
		framer.Reader(os.Stdin),
		framer.Writer(os.Stdout),
		os.Stdin,
	)
	conn.Go(ctx, handler)
	<-conn.Done()
}

// handler handles incoming requests. The context originates from the ctx
// passed to conn.Go and can be used for cancellation or deadlines in
// long-running operations.
func handler(_ context.Context, req jsonrpc2.Request) (any, error) {
	switch req.Method() {
	case "greet":
		var name string
		if err := json.Unmarshal(req.Params(), &name); err != nil {
			return nil, fmt.Errorf("%w: %w", jsonrpc2.ErrInvalidParams, err)
		}
		return fmt.Sprintf("Hello, %s!", name), nil

	case "add":
		var nums [2]float64
		if err := json.Unmarshal(req.Params(), &nums); err != nil {
			return nil, fmt.Errorf("%w: %w", jsonrpc2.ErrInvalidParams, err)
		}
		return nums[0] + nums[1], nil

	default:
		return nil, fmt.Errorf("%q: %w", req.Method(), jsonrpc2.ErrMethodNotFound)
	}
}
