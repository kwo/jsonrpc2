# jsonrpc2

Package jsonrpc2 is an implementation of the JSON-RPC 2 specification for Go.

## Features Comparison

This library is a fork of [go.lsp.dev/jsonrpc2](https://github.com/go-language-server/jsonrpc2), which itself is a cleaned-up fork of [x/tools/internal/jsonrpc2](https://github.com/golang/tools/tree/master/internal/jsonrpc2). The goal is to provide an importable, stdlib-only JSON-RPC 2 library that cherry-picks the best ideas from the Go team's v2 rewrite without the complexity.

| Feature | kwo/jsonrpc2 | go.lsp.dev/jsonrpc2 | x/tools jsonrpc2 (v1) | x/tools jsonrpc2_v2 |
|---|:---:|:---:|:---:|:---:|
| Importable | ✅ | ✅ | ❌ `internal` | ❌ `internal` |
| Stdlib only (no external deps) | ✅ | ❌ `segmentio/encoding` | ❌ `internal/event` | ✅ |
| LSP header framing | ✅ | ✅ | ✅ | ✅ |
| Raw JSON framing | ✅ | ✅ | ✅ | ✅ |
| Bidirectional Conn | ✅ | ✅ | ✅ | ✅ |
| Call / Notify / Response messages | ✅ | ✅ | ✅ | ✅ |
| Structured JSON-RPC error codes | ✅ | ✅ | ✅ | ✅ |
| Context cancellation | ✅ | ✅ | ✅ | ✅ |
| Handler middleware (cancel, async, reply) | ✅ | ✅ | ✅ | — |
| CancelHandler | ✅ | ✅ | ✅ | — |
| AsyncHandler | ✅ | ✅ | ✅ | — |
| ReplyHandler (must-reply enforcement) | ✅ | ✅ | ✅ | — |
| StreamServer / Serve / ListenAndServe | ✅ | ✅ | ✅ | — |
| Server idle timeout | ✅ | ✅ | ✅ | ✅ |
| Test helpers (TCP + pipe servers) | ✅ | ✅ | ✅ | — |
| `io.ReadWriteCloser`-based stream | ✅ | ✅ | ❌ `net.Conn` | ✅ |
| Split Reader/Writer interfaces | — | — | — | ✅ |
| AsyncCall (non-blocking Call) | — | — | — | ✅ |
| Write error propagation | — | — | — | ✅ |
| Graceful close (drain in-flight) | — | — | — | ✅ |
| Handler returns `(result, error)` | — | — | — | ✅ |
| Binder (per-connection config) | — | — | — | ✅ |
| Preempter (pre-queue handling) | — | — | — | ✅ |
| ErrAsyncResponse (deferred replies) | — | — | — | ✅ |
| IdleListener with finalizers | — | — | — | ✅ |
| Built-in event tracing/metrics | — | — | ✅ | — |
| Server.Shutdown (graceful stop) | — | — | — | ✅ |
| Listener/Dialer interfaces | — | — | — | ✅ |

## References

- [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)
- [golang.org/x/tools/internal/jsonrpc2](https://github.com/golang/tools/tree/master/internal/jsonrpc2)
- [github.com/sourcegraph/jsonrpc2](https://github.com/sourcegraph/jsonrpc2)

## OpenRPC

The OpenRPC Specification defines a standard, programming language-agnostic interface description for JSON-RPC 2.0 APIs.

- [open-rpc.org](https://open-rpc.org)
- [OpenRPC Specification](https://spec.open-rpc.org/)
- [OpenRPC GitHub organization](https://github.com/open-rpc)
