package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/microfactory/line"
)

var gatewayHandler = line.NewGatewayHandler(0, http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		sess := line.SessionFromContext(r.Context())
		db := dynamodb.New(sess)
		//@TODO how we get a table name

		_ = db

		fmt.Fprintln(w, `{"hello": "world"}`)
	}))

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
	mux.Use(line.WithSession())      //adds an aws session to the context

	return mux.Handle(msg, invoc)
}
