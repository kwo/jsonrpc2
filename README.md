# jsonrpc2

Package jsonrpc2 is an implementation of the JSON-RPC 2 specification for Go.

## Features Comparison

This library is a fork of [go.lsp.dev/jsonrpc2](https://github.com/go-language-server/jsonrpc2), which itself is a cleaned-up fork of [x/tools/internal/jsonrpc2](https://github.com/golang/tools/tree/master/internal/jsonrpc2). The goal is to provide an importable, stdlib-only JSON-RPC 2 library that cherry-picks the best ideas from the Go team's v2 rewrite without the complexity.

| Feature | kwo/jsonrpc2 | go.lsp.dev/jsonrpc2 | x/tools jsonrpc2 (v1) | x/tools jsonrpc2_v2 |
|---|:---:|:---:|:---:|:---:|
| [Importable](## "Can be imported by external Go modules") | ✅ | ✅ | ❌ `internal` | ❌ `internal` |
| [Stdlib only (no external deps)](## "Uses only the Go standard library, no third-party dependencies") | ✅ | ❌ `segmentio/encoding` | ❌ `internal/event` | ✅ |
| [LOC (Go, non-test)](## "Approximate lines of Go source code, excluding test files") | 1,447 | 1,533 | 1,833 | 1,859 |
| [LSP header framing](## "Content-Length header framing as used by the Language Server Protocol") | ✅ | ✅ | ✅ | ✅ |
| [Raw JSON framing](## "Newline-delimited JSON with no framing headers") | ✅ | ✅ | ✅ | ✅ |
| [Bidirectional Conn](## "A single Conn can both send and receive requests") | ✅ | ✅ | ✅ | ✅ |
| [Call / Notify / Response messages](## "Typed message structs for requests expecting a reply, fire-and-forget notifications, and responses") | ✅ | ✅ | ✅ | ✅ |
| [Structured JSON-RPC error codes](## "Error type with numeric code, message, and optional data per the JSON-RPC 2.0 spec") | ✅ | ✅ | ✅ | ✅ |
| [Context cancellation](## "Calls and handlers respect context.Context for cancellation and deadlines") | ✅ | ✅ | ✅ | ✅ |
| [Server idle timeout](## "Server exits automatically after no active connections for a configured duration") | ✅ | ✅ | ✅ | ✅ |
| [Split Reader/Writer interfaces](## "Separate Reader and Writer interfaces instead of a single Stream, enabling natural stdin/stdout usage") | ✅ | — | — | ✅ |
| [AsyncCall (non-blocking Call)](## "Fire a call and get a handle to Await the result later, enabling concurrent in-flight requests") | ✅ | — | — | ✅ |
| [Write error propagation](## "A write failure is recorded and fails all subsequent writes, closing the transport to unblock pending calls") | ✅ | — | — | ✅ |
| [Handler returns `(result, error)`](## "Handlers return a value and error directly; the conn sends the reply, so the compiler enforces a response") | ✅ | — | — | ✅ |
| [Graceful close (drain in-flight)](## "Close stops accepting new calls but waits for in-flight handlers to finish before shutting down") | — | — | — | ✅ |
| [Server.Shutdown (graceful stop)](## "Explicit Shutdown method that drains connections before stopping the server") | — | — | — | ✅ |
| [CancelHandler](## "Middleware that gives each Call a cancellable context and exposes a function to cancel in-progress requests by ID") | ✅ | ✅ | ✅ | — |
| [AsyncHandler](## "Middleware that runs each request in its own goroutine, ordered by arrival, to unblock the read loop") | ✅ | ✅ | ✅ | — |
| [StreamServer / Serve / ListenAndServe](## "Server helpers that accept network connections and serve each with a StreamServer, similar to net/http") | ✅ | ✅ | ✅ | — |
| [Test helpers (TCP + pipe servers)](## "TCPServer and PipeServer helpers for writing integration tests against a jsonrpc2 server") | ✅ | ✅ | ✅ | — |
| [ReplyHandler (must-reply enforcement)](## "Middleware that panics if a handler does not call Reply exactly once; superseded by return-value Handler") | — | ✅ | ✅ | — |
| [Binder (per-connection config)](## "Callback invoked for each new connection to configure handler and options") | — | — | — | ✅ |
| [Preempter (pre-queue handling)](## "Handler that runs before the main queue, used for LSP cancel notifications") | — | — | — | ✅ |
| [ErrAsyncResponse (deferred replies)](## "Sentinel error allowing a handler to defer its reply to a later point") | — | — | — | ✅ |
| [IdleListener with finalizers](## "Listener wrapper that tracks idle time with runtime finalizers for cleanup") | — | — | — | ✅ |
| [Built-in event tracing/metrics](## "Integrated event-based tracing and metrics via internal/event") | — | — | ✅ | — |
| [Listener/Dialer interfaces](## "Abstractions over net.Listener and net.Dialer for custom transports") | — | — | — | ✅ |

LOC note: counts are approximate Go source lines for each implementation, excluding `*_test.go`. For `x/tools`, counts are for the `internal/jsonrpc2` and `internal/jsonrpc2_v2` directories only.

## References

- [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)
- [golang.org/x/tools/internal/jsonrpc2](https://github.com/golang/tools/tree/master/internal/jsonrpc2)
- [github.com/sourcegraph/jsonrpc2](https://github.com/sourcegraph/jsonrpc2)

## OpenRPC

The OpenRPC Specification defines a standard, programming language-agnostic interface description for JSON-RPC 2.0 APIs.

- [open-rpc.org](https://open-rpc.org)
- [OpenRPC Specification](https://spec.open-rpc.org/)
- [OpenRPC GitHub organization](https://github.com/open-rpc)
