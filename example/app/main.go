// Command client starts the server process and demonstrates how to make
// JSON-RPC 2.0 calls over stdin/stdout pipes.
//
// Run from the example/ directory:
//
//	go run ./client
package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/kwo/jsonrpc2"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Start the plugin as a child process.
	cmd := exec.Command("go", "run", "./plugin")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Connect to the plugin over its stdin/stdout.
	framer := jsonrpc2.RawFramer()
	conn := jsonrpc2.NewConn(
		framer.Reader(stdout),
		framer.Writer(stdin),
		stdin,
	)
	// The connection is bidirectional: replace MethodNotFoundHandler with a
	// real handler to receive calls and notifications from the plugin.
	conn.Go(ctx, jsonrpc2.MethodNotFoundHandler)

	// --- Synchronous call ---
	var greeting string
	if _, err := conn.Call(ctx, "greet", "World", &greeting); err != nil {
		log.Fatal(err)
	}
	fmt.Println(greeting)

	// --- Async calls (concurrent) ---
	ac1, err := conn.AsyncCall(ctx, "add", [2]float64{2, 3})
	if err != nil {
		log.Fatal(err)
	}
	ac2, err := conn.AsyncCall(ctx, "add", [2]float64{10, 20})
	if err != nil {
		log.Fatal(err)
	}

	var sum1, sum2 float64
	if err := ac1.Await(ctx, &sum1); err != nil {
		log.Fatal(err)
	}
	if err := ac2.Await(ctx, &sum2); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("2 + 3 = %.0f\n", sum1)
	fmt.Printf("10 + 20 = %.0f\n", sum2)

	// --- Notification (fire-and-forget, no response) ---
	if err := conn.Notify(ctx, "greet", "nobody"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("notification sent")

	// Clean up.
	conn.Close()
	<-conn.Done()
	_ = cmd.Wait()
}
