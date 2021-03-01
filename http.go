package bug

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/felixge/httpsnoop"
)

// HTTPResponseMiddleware returns an HTTP handler that logs the response.
func HTTPResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span, cancel := WithSpan(r.Context(), "response")
		defer cancel()

		span.Append(
			Tag("method", r.Method),
			Tag("path", r.URL.Path),
			Tag("remote-addr", r.RemoteAddr))

		m := httpsnoop.CaptureMetrics(next, w, r.WithContext(span))

		span.Append(
			Tag("body", m.Written),
			Tag("status", m.Code))
	})

}

// HTTPPanicMiddleware returns an HTTP handler that recovers from a panic caused by the
// next handler. It logs a panic event including the value, type, and source location.
// After logging, the handler tries to respond with an internal server error.
func HTTPPanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var returned bool

		defer func() {
			if v := recover(); !returned {
				pc := make([]uintptr, 64)
				n := runtime.Callers(3, pc)
				frames := runtime.CallersFrames(pc[:n])

				var fn string
				var stack []string

				for {
					frame, more := frames.Next()
					stack = append(stack, fmt.Sprintf("%s:%d", frame.File, frame.Line))

					if f := frame.Func; f != nil && fn == "" {
						fn = f.Name()
					}

					if !more {
						break
					}
				}

				taggers := []Tagger{
					Tag("func", fn),
					Tag("stack", stack),
					Tag("type", fmt.Sprintf("%T", v)),
				}

				if err, ok := v.(error); ok {
					v = err.Error()
				}

				taggers = append(taggers,
					Tag("value", v))

				Log(r.Context(), "panic", taggers...)

				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("bug: Internal server error\n"))
			}
		}()

		next.ServeHTTP(w, r)
		returned = true
	})
}
