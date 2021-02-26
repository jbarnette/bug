# bug, a logger

bug is a simple & slow structured logger for context-aware Go programs. It writes JSONL to stdout by default. To log an event, call `bug.Log`. Use `bug.With` to add tags to the context. Use `bug.WithSpan` to measure the duration of a task.

```go
package main

import (
  "context"
  "time"

  "github.com/jbarnette/bug"
)

func main() {
  ctx := context.Background()

  // log an event
  bug.Log(ctx, "hello",
    bug.Tag("world", true))

  // add shared tags
  ctx = bug.With(ctx,
    bug.Tag("shared", "value"))

  // measure elapsed time
  ctx, cancel := bug.WithSpan(ctx, "leaving")
  defer cancel()

  time.Sleep(1 * time.Second)
  bug.Log(ctx, "goodbye")
}
```

```json
{"at":"hello","world":true}
{"at":"goodbye","shared":"value"}
{"at":"leaving","elapsed":1.00043627,"shared":"value"}
```

## Error tags

`bug.Error` tags an event with the error's message and type. It ignores nil values.

```go
bug.Log(ctx, "accept",
  bug.Error(err))
```

```json
{"at":"accept","error":true,"error.message":"accept tcp 127.0.0.1:32827: use of closed network connection","error.type":"*net.OpError"}
```

## Logfmt

You can filter bug's output through [jbarnette/logfmt](https://github.com/jbarnette/logfmt) for a more humane view.
