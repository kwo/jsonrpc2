// SPDX-FileCopyrightText: 2021 The Go Language Server Authors
// SPDX-FileCopyrightText: 2026 Karl Ostendorf
// SPDX-License-Identifier: BSD-3-Clause

package fake_test

import (
	"context"
	"testing"
	"time"

	"github.com/kwo/jsonrpc2"
	"github.com/kwo/jsonrpc2/fake"
)

type msg struct {
	Msg string
}

func fakeHandler(_ context.Context, _ jsonrpc2.Request) (interface{}, error) {
	return &msg{"pong"}, nil
}

func TestTestServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	server := jsonrpc2.HandlerServer(fakeHandler)

	tcpTS := fake.NewTCPServer(ctx, server, nil)
	defer func() {
		if err := tcpTS.Close(); err != nil {
			t.Error(err)
		}
	}()

	pipeTS := fake.NewPipeServer(ctx, server, nil)
	defer func() {
		if err := pipeTS.Close(); err != nil {
			t.Error(err)
		}
	}()

	tests := map[string]struct {
		connector fake.Connector
	}{
		"tcp": {
			connector: tcpTS,
		},
		"pipe": {
			connector: pipeTS,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			conn := tt.connector.Connect(ctx)
			conn.Go(ctx, jsonrpc2.MethodNotFoundHandler)

			var got msg
			if _, err := conn.Call(ctx, "ping", &msg{"ping"}, &got); err != nil {
				t.Fatal(err)
			}

			if want := "pong"; got.Msg != want {
				t.Errorf("conn.Call(...): returned %q, want %q", got, want)
			}
		})
	}
}
