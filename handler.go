// SPDX-FileCopyrightText: 2019 The Go Language Server Authors
// SPDX-FileCopyrightText: 2019 The Go Authors
// SPDX-FileCopyrightText: 2026 Karl Ostendorf
// SPDX-License-Identifier: BSD-3-Clause

package jsonrpc2

import (
	"context"
	"fmt"
	"sync"
)

// Handler is invoked to handle incoming requests.
//
// Returning a non-nil error for a Call will send an error response.
// Returning a non-nil result will send a success response.
// For Notifications, the return values are ignored.
type Handler func(ctx context.Context, req Request) (result interface{}, err error)

// MethodNotFoundHandler is a Handler that replies to all call requests with the
// standard method not found response.
//
// This should normally be the final handler in a chain.
func MethodNotFoundHandler(_ context.Context, req Request) (interface{}, error) {
	return nil, fmt.Errorf("%q: %w", req.Method(), ErrMethodNotFound)
}

// CancelHandler returns a handler that supports cancellation, and a function
// that can be used to trigger canceling in progress requests.
func CancelHandler(handler Handler) (Handler, func(id ID)) {
	var mu sync.Mutex
	handling := make(map[ID]context.CancelFunc)

	h := Handler(func(ctx context.Context, req Request) (interface{}, error) {
		if call, ok := req.(*Call); ok {
			cancelCtx, cancel := context.WithCancel(ctx)
			mu.Lock()
			handling[call.ID()] = cancel
			mu.Unlock()
			defer func() {
				mu.Lock()
				delete(handling, call.ID())
				mu.Unlock()
				cancel()
			}()
			return handler(cancelCtx, req)
		}
		return handler(ctx, req)
	})

	canceller := func(id ID) {
		mu.Lock()
		cancel, found := handling[id]
		mu.Unlock()
		if found {
			cancel()
		}
	}

	return h, canceller
}

// AsyncHandler returns a handler that processes each request in its own
// goroutine.
//
// Each request waits for the previous request to finish before it starts.
// This allows the read loop to unblock at the cost of unbounded goroutines
// all stalled on the previous one.
func AsyncHandler(handler Handler) Handler {
	nextRequest := make(chan struct{})
	close(nextRequest)

	return func(ctx context.Context, req Request) (interface{}, error) {
		waitForPrevious := nextRequest
		nextRequest = make(chan struct{})
		unlockNext := nextRequest

		type result struct {
			val interface{}
			err error
		}
		done := make(chan result, 1)

		go func() {
			<-waitForPrevious
			val, err := handler(ctx, req)
			close(unlockNext)
			done <- result{val, err}
		}()

		r := <-done
		return r.val, r.err
	}
}
