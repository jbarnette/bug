# bug, a logger

bug is a simple & slow structured logger for context-aware Go programs. It writes JSONL to stdout by default. To log an event, call `bug.Log`. Use `bug.With` to add tags to the context. Measure elapsed time during an operation with `bug.Span`.

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
  ctx, cancel := bug.Span(ctx, "leaving")
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

## Logfmt

You can filter bug's output through [jbarnette/logfmt](https://github.com/jbarnette/logfmt) for a more humane view.
