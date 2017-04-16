package main

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/microfactory/line"
	"github.com/microfactory/line/session"
)

var gatewayHandler = line.HandlerFunc(
	func(ctx context.Context, msg json.RawMessage) (interface{}, error) {
		return "hello gateway", nil
	})

var scheduleHandler = line.HandlerFunc(
	func(ctx context.Context, msg json.RawMessage) (interface{}, error) {
		return "hello schedule", nil
	})

//Handle will route lambda events to the specific handler
func Handle(msg json.RawMessage, invoc *line.Invocation) (interface{}, error) {
	mux := line.NewMux()
	mux.MatchARN(regexp.MustCompile(`-gateway$`), gatewayHandler)
	mux.MatchARN(regexp.MustCompile(`-schedule$`), gatewayHandler)

	mux.Use(line.EarlyTimeout(5000)) //gives handlers 5sec to clean up
	mux.Use(session.LineSession())   //add an aws session to the context

	return mux.Handle(msg, invoc)
}
