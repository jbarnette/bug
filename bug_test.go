package bug_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jbarnette/bug"
)

func ExampleLog() {
	bug.Write = bug.JSONL(os.Stdout)
	ctx := context.Background()

	bug.Log(ctx, "hello",
		bug.Tag("greeting", true),
		bug.Tag("subject", "world"))

	// Output:
	// {"at":"hello","greeting":true,"subject":"world"}
}

func ExampleError() {
	bug.Write = bug.JSONL(os.Stdout)
	ctx := context.Background()

	bug.Log(ctx, "no-boom",
		bug.Error(nil))

	err := errors.New("kaboom")

	bug.Log(ctx, "boom",
		bug.Error(err))

	// Output:
	// {"at":"no-boom"}
	// {"at":"boom","error":true,"error.message":"kaboom","error.type":"*errors.errorString"}
}

func ExampleWith() {
	bug.Write = bug.JSONL(os.Stdout)
	ctx := context.Background()

	bug.Log(ctx, "with-background")

	ctx = bug.With(ctx,
		bug.Tag("outer", "value"))

	bug.Log(ctx, "with-outer")

	ctx = bug.With(ctx,
		bug.Tag("inner", "value"))

	bug.Log(ctx, "with-inner",
		bug.Tag("inline", "value"))

	// Output:
	// {"at":"with-background"}
	// {"at":"with-outer","outer":"value"}
	// {"at":"with-inner","inline":"value","inner":"value","outer":"value"}
}

func ExampleWithSpan() {
	bug.Write = bug.JSONL(os.Stdout)
	ctx := context.Background()

	t := &time.Time{}
	bug.Now = func() time.Time { return *t }

	span, cancel := bug.WithSpan(ctx, "outside")
	defer cancel()

	span.Append(bug.Tag("outer", "value"))
	bug.Log(span, "inside", bug.Tag("inner", "value"))

	*t = t.Add(1 * time.Minute)

	// Output:
	// {"at":"inside","inner":"value"}
	// {"at":"outside","elapsed":60,"outer":"value"}
}

func TestSpanFrom(t *testing.T) {
	ctx := context.Background()
	span, cancel := bug.WithSpan(ctx, "testing")
	defer cancel()

	if from := bug.SpanFrom(span); from != span {
		t.Errorf("%v != %v", from, span)
	}
}

type stringer struct{ string }

func (s stringer) String() string {
	return s.string
}

func ExampleJSONL() {
	bug.Write = bug.JSONL(os.Stdout)
	ctx := context.Background()

	bug.Log(ctx, "func",
		bug.Tag("value", func() {}))

	bug.Log(ctx, "chan",
		bug.Tag("value", make(chan struct{})))

	bug.Log(ctx, "stringer",
		bug.Tag("value", stringer{"value"}))

	// Output:
	// {"at":"func","value":"💥"}
	// {"at":"chan","value":"💥"}
	// {"at":"stringer","value":"value"}
}
