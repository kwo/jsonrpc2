# Example: App ↔ Plugin over stdin/stdout

This example demonstrates bidirectional JSON-RPC 2.0 communication between
two processes connected via stdin/stdout pipes.

- `plugin/` — a JSON-RPC server that handles `greet` and `add` methods
- `app/` — starts the plugin as a child process and makes calls to it

## Running

From this directory:

```sh
go run ./app
```

That's it — the app starts the plugin automatically. You should see:

```
Hello, World!
2 + 3 = 5
10 + 20 = 30
notification sent
```

There is no need to start the plugin manually. The app launches it as a child process and connects over its stdin/stdout using `RawFramer`.
