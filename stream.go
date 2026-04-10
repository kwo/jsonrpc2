// SPDX-FileCopyrightText: 2018 The Go Language Server Authors
// SPDX-FileCopyrightText: 2019 The Go Authors
// SPDX-FileCopyrightText: 2026 Karl Ostendorf
// SPDX-License-Identifier: BSD-3-Clause

package jsonrpc2

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	// HdrContentLength is the HTTP header name of the length of the content part in bytes. This header is required.
	//
	// RFC 7230, section 3.3.2: Content-Length:
	//  https://tools.ietf.org/html/rfc7230#section-3.3.2
	HdrContentLength = "Content-Length"

	// HeaderContentType is the mime type of the content part. Defaults to "application/vscode-jsonrpc; charset=utf-8".
	//
	// RFC 7231, section 3.1.1.5: Content-Type:
	//  https://tools.ietf.org/html/rfc7231#section-3.1.1.5
	HdrContentType = "Content-Type"

	// HeaderContentSeparator is the header and content part separator.
	HdrContentSeparator = "\r\n\r\n"
)

// Reader reads JSON-RPC messages.
type Reader interface {
	Read(context.Context) (Message, int64, error)
}

// Writer writes JSON-RPC messages.
type Writer interface {
	Write(context.Context, Message) (int64, error)
}

// Framer produces Readers and Writers for a given transport.
type Framer interface {
	Reader(io.Reader) Reader
	Writer(io.Writer) Writer
}

// HeaderFramer returns a Framer that uses LSP-style Content-Length headers.
func HeaderFramer() Framer { return headerFramer{} }

// RawFramer returns a Framer that uses raw JSON with no framing headers.
func RawFramer() Framer { return rawFramer{} }

// headerFramer implements Framer for LSP-style header framing.
type headerFramer struct{}

func (headerFramer) Reader(r io.Reader) Reader { return &headerReader{in: bufio.NewReader(r)} }
func (headerFramer) Writer(w io.Writer) Writer  { return &headerWriter{out: w} }

// rawFramer implements Framer for raw JSON framing.
type rawFramer struct{}

func (rawFramer) Reader(r io.Reader) Reader { return &rawReader{in: json.NewDecoder(r)} }
func (rawFramer) Writer(w io.Writer) Writer  { return &rawWriter{out: w} }

// rawReader reads raw JSON messages.
type rawReader struct {
	in *json.Decoder
}

func (r *rawReader) Read(ctx context.Context) (Message, int64, error) {
	select {
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	default:
	}

	var raw json.RawMessage
	if err := r.in.Decode(&raw); err != nil {
		return nil, 0, fmt.Errorf("decoding raw message: %w", err)
	}

	msg, err := DecodeMessage(raw)
	return msg, int64(len(raw)), err
}

// rawWriter writes raw JSON messages.
type rawWriter struct {
	out io.Writer
}

func (w *rawWriter) Write(ctx context.Context, msg Message) (int64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return 0, fmt.Errorf("marshaling message: %w", err)
	}

	n, err := w.out.Write(data)
	if err != nil {
		return 0, fmt.Errorf("write to stream: %w", err)
	}

	return int64(n), nil
}

// headerReader reads LSP-framed messages.
type headerReader struct {
	in *bufio.Reader
}

func (r *headerReader) Read(ctx context.Context) (Message, int64, error) {
	select {
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	default:
	}

	var total, length int64
	for {
		line, err := r.in.ReadString('\n')
		total += int64(len(line))
		if err != nil {
			return nil, total, fmt.Errorf("failed reading header line: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		colon := strings.IndexRune(line, ':')
		if colon < 0 {
			return nil, total, fmt.Errorf("invalid header line %q", line)
		}

		name, value := line[:colon], strings.TrimSpace(line[colon+1:])
		switch name {
		case HdrContentLength:
			if length, err = strconv.ParseInt(value, 10, 32); err != nil {
				return nil, total, fmt.Errorf("failed parsing %s: %v: %w", HdrContentLength, value, err)
			}
			if length <= 0 {
				return nil, total, fmt.Errorf("invalid %s: %v", HdrContentLength, length)
			}
		default:
			// ignoring unknown headers
		}
	}

	if length == 0 {
		return nil, total, fmt.Errorf("missing %s header", HdrContentLength)
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(r.in, data); err != nil {
		return nil, total, fmt.Errorf("read full of data: %w", err)
	}

	total += length
	msg, err := DecodeMessage(data)
	return msg, total, err
}

// headerWriter writes LSP-framed messages.
type headerWriter struct {
	out io.Writer
}

func (w *headerWriter) Write(ctx context.Context, msg Message) (int64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return 0, fmt.Errorf("marshaling message: %w", err)
	}

	n, err := fmt.Fprintf(w.out, "%s: %v%s", HdrContentLength, len(data), HdrContentSeparator)
	total := int64(n)
	if err != nil {
		return 0, fmt.Errorf("write data to conn: %w", err)
	}

	n, err = w.out.Write(data)
	total += int64(n)
	if err != nil {
		return 0, fmt.Errorf("write data to conn: %w", err)
	}

	return total, nil
}
