// Package bug provides a simple & slow structured logger. It writes JSONL to stdout.
package bug

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// A Span is a context that covers the lifetime of an operation.
type Span struct {
	context.Context
	taggers []Tagger
}

// A Tagger is a generator function for log event tags.
type Tagger func(tag func(key string, value any))

// A Writer is a function that generates and writes a log event.
type Writer func(ctx context.Context, at string, tagger Tagger)

type any = interface{}

type spanKey struct{}

type taggerKey struct{}

// Write is called by Log to actually log an event.
var Write Writer = JSONL(os.Stdout)

// Now is called by Measure to measure elapsed time.
var Now func() time.Time = time.Now

// jsonPlaceholder is used when the JSONL writer can't marshal a tagged value.
var jsonPlaceholder = []byte(`"💥"`)

// Log calls Write to write a log event.
func Log(ctx context.Context, at string, taggers ...Tagger) {
	Write(ctx, at, combine(ctx, taggers))
}

// Tag returns a tagger for the provided key and value. Pass it to Log or With.
func Tag(key string, value any) Tagger {
	return func(tag func(string, any)) {
		tag(key, value)
	}
}

// Error returns a tagger for the provided error. If the error is non-nil, the tagger
// generates error=true, error.message, and error.type tags.
func Error(err error) Tagger {
	return func(tag func(string, any)) {
		if err == nil {
			return
		}

		tag("error", true)
		tag("error.message", err.Error())
		tag("error.type", fmt.Sprintf("%T", err))
	}
}

// With returns a copy of parent with additional taggers.
func With(parent context.Context, taggers ...Tagger) context.Context {
	return context.WithValue(parent, taggerKey{}, combine(parent, taggers))
}

// WithSpan returns a copy of parent with a CancelFunc that calls Log. The CancelFunc adds
// an "elapsed" tag, calculated by subtracting Now from the time when Span was called.
func WithSpan(parent context.Context, at string) (*Span, context.CancelFunc) {
	start := Now()

	ctx, cancel := context.WithCancel(parent)
	span := &Span{Context: ctx}

	log := func() {
		cancel()

		elapsed := Now().Sub(start)
		span.Append(Tag("elapsed", elapsed.Seconds()))
		Log(parent, at, span.taggers...)
	}

	return span, log
}

// SpanFrom returns a pointer to the current span, if there is one.
func SpanFrom(ctx context.Context) *Span {
	if span, ok := ctx.Value(spanKey{}).(*Span); ok {
		return span
	}

	return nil
}

// JSONL returns a log writer that writes JSON lines to w.
func JSONL(w io.Writer) Writer {
	mu := &sync.Mutex{}
	en := json.NewEncoder(w)

	return func(ctx context.Context, at string, tag Tagger) {
		mu.Lock()
		defer mu.Unlock()

		e := map[string]json.RawMessage{
			"at": safeJSONMarshal(at),
		}

		tag(func(k string, v any) {
			e[k] = safeJSONMarshal(v)
		})

		if err := en.Encode(e); err != nil {
			fmt.Fprintf(os.Stderr, "bug: jsonl: %v\n", err)
		}
	}
}

// Append appends the provided taggers to the span's taggers.
func (s *Span) Append(taggers ...Tagger) {
	s.taggers = append(s.taggers, taggers...)
}

// Value implements context.Context with a special case to support SpanFrom.
func (s *Span) Value(key interface{}) interface{} {
	if key == (spanKey{}) {
		return s
	}

	return s.Context.Value(key)
}

// combine returns a Tagger that combines ctx's tagger with the provided taggers.
func combine(ctx context.Context, taggers []Tagger) Tagger {
	return func(f func(string, any)) {
		if old, ok := ctx.Value(taggerKey{}).(Tagger); ok {
			old(f)
		}

		for _, t := range taggers {
			t(f)
		}
	}
}

// safeJSONMarshal marshals value to JSON. It returns a placeholder if it encounters an error.
// If the value is a fmt.Stringer, the result of value.String is marshaled.
func safeJSONMarshal(value any) json.RawMessage {
	if s, ok := value.(fmt.Stringer); ok {
		value = s.String()
	}

	if raw, err := json.Marshal(value); err == nil {
		return raw
	}

	return jsonPlaceholder
}
