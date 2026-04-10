// SPDX-FileCopyrightText: 2021 The Go Language Server Authors
// SPDX-License-Identifier: BSD-3-Clause

package jsonrpc2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

// Conn is the common interface to jsonrpc clients and servers.
//
// Conn is bidirectional; it does not have a designated server or client end.
// It manages the jsonrpc2 protocol, connecting responses back to their calls.
type Conn interface {
	// Call invokes the target method and waits for a response.
	// The response will be unmarshaled from JSON into the result.
	Call(ctx context.Context, method string, params, result interface{}) (ID, error)

	// AsyncCall starts a call to the target method but does not wait for the
	// response. The returned AsyncCall can be used to await the result later.
	// This allows multiple calls to be in flight concurrently.
	AsyncCall(ctx context.Context, method string, params interface{}) (*AsyncCall, error)

	// Notify invokes the target method but does not wait for a response.
	Notify(ctx context.Context, method string, params interface{}) error

	// Go starts a goroutine to handle the connection.
	//
	// It must be called exactly once for each Conn. It returns immediately.
	// Must block on Done() to wait for the connection to shut down.
	//
	// This is a temporary measure, this should be started automatically in the
	// future.
	Go(ctx context.Context, handler Handler)

	// Close closes the connection and its underlying stream.
	//
	// It does not wait for the close to complete, use the Done() channel for
	// that.
	Close() error

	// Done returns a channel that will be closed when the processing goroutine
	// has terminated, which will happen if Close() is called or an underlying
	// stream is closed.
	Done() <-chan struct{}

	// Err returns an error if there was one from within the processing goroutine.
	//
	// If err returns non nil, the connection will be already closed or closing.
	Err() error
}

type conn struct {
	seq       int32              // access atomically
	writeMu   sync.Mutex         // protects writes to the writer
	reader    Reader             // reads messages
	writer    Writer             // writes messages
	closer    io.Closer          // closes the underlying transport
	pendingMu sync.Mutex         // protects the pending map
	pending   map[ID]chan *Response

	done chan struct{}
	err  atomic.Value
}

// NewConn creates a new connection object around the supplied reader, writer,
// and closer.
func NewConn(r Reader, w Writer, c io.Closer) Conn {
	return &conn{
		reader:  r,
		writer:  w,
		closer:  c,
		pending: make(map[ID]chan *Response),
		done:    make(chan struct{}),
	}
}

// AsyncCall represents an in-flight call that has been sent but whose
// response has not yet been received.
type AsyncCall struct {
	id    ID
	rchan <-chan *Response
	conn  *conn
}

// ID returns the request ID of this call.
func (ac *AsyncCall) ID() ID { return ac.id }

// Await waits for the response and unmarshals it into result.
func (ac *AsyncCall) Await(ctx context.Context, result interface{}) error {
	select {
	case resp := <-ac.rchan:
		ac.conn.removePending(ac.id)
		if resp.err != nil {
			return resp.err
		}
		if result == nil || len(resp.result) == 0 {
			return nil
		}
		if err := json.Unmarshal(resp.result, result); err != nil {
			return fmt.Errorf("unmarshaling result: %w", err)
		}
		return nil
	case <-ctx.Done():
		ac.conn.removePending(ac.id)
		return ctx.Err()
	}
}

// Call implements Conn.
func (c *conn) Call(ctx context.Context, method string, params, result interface{}) (ID, error) {
	ac, err := c.AsyncCall(ctx, method, params)
	if err != nil {
		return ac.ID(), err
	}
	return ac.ID(), ac.Await(ctx, result)
}

// AsyncCall implements Conn.
func (c *conn) AsyncCall(ctx context.Context, method string, params interface{}) (*AsyncCall, error) {
	id := NewNumberID(atomic.AddInt32(&c.seq, 1))
	call, err := NewCall(id, method, params)
	if err != nil {
		return &AsyncCall{id: id}, fmt.Errorf("marshaling call parameters: %w", err)
	}

	rchan := make(chan *Response, 1)

	c.pendingMu.Lock()
	c.pending[id] = rchan
	c.pendingMu.Unlock()

	_, err = c.write(ctx, call)
	if err != nil {
		c.removePending(id)
		return &AsyncCall{id: id}, err
	}

	return &AsyncCall{id: id, rchan: rchan, conn: c}, nil
}

func (c *conn) removePending(id ID) {
	c.pendingMu.Lock()
	delete(c.pending, id)
	c.pendingMu.Unlock()
}

// Notify implements Conn.
func (c *conn) Notify(ctx context.Context, method string, params interface{}) error {
	notify, err := NewNotification(method, params)
	if err != nil {
		return fmt.Errorf("marshaling notify parameters: %w", err)
	}

	_, err = c.write(ctx, notify)
	return err
}

func (c *conn) replier(req Message) Replier {
	return func(ctx context.Context, result interface{}, err error) error {
		call, ok := req.(*Call)
		if !ok {
			return nil
		}

		response, err := NewResponse(call.id, result, err)
		if err != nil {
			return err
		}

		_, err = c.write(ctx, response)
		if err != nil {
			// TODO(iancottrell): if a stream write fails, we really need to shut down the whole stream
			return err
		}
		return nil
	}
}

func (c *conn) write(ctx context.Context, msg Message) (int64, error) {
	c.writeMu.Lock()
	n, err := c.writer.Write(ctx, msg)
	c.writeMu.Unlock()
	if err != nil {
		return 0, fmt.Errorf("write to stream: %w", err)
	}

	return n, nil
}

// Go implements Conn.
func (c *conn) Go(ctx context.Context, handler Handler) {
	go c.run(ctx, handler)
}

func (c *conn) run(ctx context.Context, handler Handler) {
	defer close(c.done)

	for {
		msg, _, err := c.reader.Read(ctx)
		if err != nil {
			c.fail(err)
			return
		}

		switch msg := msg.(type) {
		case Request:
			if err := handler(ctx, c.replier(msg), msg); err != nil {
				c.fail(err)
			}

		case *Response:
			c.pendingMu.Lock()
			rchan, ok := c.pending[msg.id]
			c.pendingMu.Unlock()
			if ok {
				rchan <- msg
			}
		}
	}
}

// Close implements Conn.
func (c *conn) Close() error {
	return c.closer.Close()
}

// Done implements Conn.
func (c *conn) Done() <-chan struct{} {
	return c.done
}

// Err implements Conn.
func (c *conn) Err() error {
	if err := c.err.Load(); err != nil {
		return err.(error)
	}
	return nil
}

func (c *conn) fail(err error) {
	c.err.Store(errors.Join(err, c.closer.Close()))
}
