package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/advanderveer/go-dynamo"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/microfactory/line"
)

var gatewayHandler = line.NewGatewayHandler(0, http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		sess := line.RuntimeSession(r.Context())
		tableName := line.ResourceAttribute(r.Context(), "my-table-name")
		db := dynamodb.New(sess)

		items := []map[string]interface{}{}
		inp := dynamo.NewScanInput(tableName)
		if _, err := dynamo.Scan(db, inp, &items); err != nil {
			fmt.Fprintf(w, `{"message": "%s"}`, err.Error())
			return
		}

		enc := json.NewEncoder(w)
		err := enc.Encode(items)
		if err != nil {
			fmt.Fprintf(w, `{"message": "%s"}`, err.Error())
			return
		}
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

	mux.Use(line.EarlyTimeout(5000))   //gives handlers 5sec to clean up
	mux.Use(line.WithRuntimeSession()) //adds an aws session to the context
	mux.Use(line.ResourceAttributes()) //adds Terraform resource attributes

	return mux.Handle(msg, invoc)
}
