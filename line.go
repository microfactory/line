package line

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

//Handler interface allows any type to handle lambda events
type Handler interface {
	HandleEvent(ctx context.Context, msg json.RawMessage) (interface{}, error)
}

//HandlerFunc is an adaptor that can handle
type HandlerFunc func(ctx context.Context, msg json.RawMessage) (interface{}, error)

//Middleware allows plugins to manipulate the context and message passed to handlers
type Middleware func(Handler) Handler

func buildChain(f Handler, m ...Middleware) Handler {
	if len(m) == 0 {
		return f
	}

	return m[0](buildChain(f, m[1:cap(m)]...))
}

//HandleEvent implements Handler
func (h HandlerFunc) HandleEvent(ctx context.Context, msg json.RawMessage) (interface{}, error) {
	return h(ctx, msg)
}

//Mux mutiplexes events to more specific handlers based on regexp matching of a context field
type Mux struct {
	handlers   map[*regexp.Regexp]Handler
	middleware []Middleware
}

//NewMux sets up new event multiplexer
func NewMux() *Mux {
	return &Mux{
		handlers: map[*regexp.Regexp]Handler{},
	}
}

//MatchARN adds a handler to the multiplexer that will be called when a caller matches the ARN of the invoked lambda function
func (mux *Mux) MatchARN(exp *regexp.Regexp, h Handler) {
	mux.handlers[exp] = h
}

//Use adds a middleware to the lambda handler
func (mux *Mux) Use(mw Middleware) {
	mux.middleware = append(mux.middleware, mw)
}

//Handle mill match the invoked function arn to a specific handler
func (mux *Mux) Handle(msg json.RawMessage, invoc *Invocation) (interface{}, error) {
	var testedExp []string
	for exp, handler := range mux.handlers {
		if exp.MatchString(invoc.InvokedFunctionARN) {
			ctx := context.Background()
			ctx = context.WithValue(ctx, invocationKey, invoc)

			wrapped := buildChain(HandlerFunc(func(ctx context.Context, msg json.RawMessage) (interface{}, error) {
				return handler.HandleEvent(ctx, msg)
			}), mux.middleware...)

			return wrapped.HandleEvent(ctx, msg)
		}

		testedExp = append(testedExp, exp.String())
	}

	return nil, fmt.Errorf("none of the handlers (%s) matched the invoked function's ARN '%s'", strings.Join(testedExp, ", "), invoc.InvokedFunctionARN)
}
