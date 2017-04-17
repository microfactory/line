package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/advanderveer/go-dynamo"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/microfactory/line"
)

//GatewayHandler handles events from the API gateway stream, in this hello world exeample it simply returns a scan of the table referenced through Terraform as 'my-table-name'
var GatewayHandler = line.NewGatewayHandler(0, http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		sess := line.RuntimeSession(r.Context())
		tname := line.ResourceAttribute(r.Context(), "my-table-name")
		db := dynamodb.New(sess)

		items := []interface{}{}
		if _, err := dynamo.NewScan(tname).Execute(db, items); err != nil {
			fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
			return
		}

		enc := json.NewEncoder(w)
		err := enc.Encode(items)
		if err != nil {
			fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
			return
		}
	}))

//Handle will route lambda events to the specific handler
func Handle(msg json.RawMessage, invoc *line.Invocation) (interface{}, error) {
	mux := line.NewMux()
	mux.MatchARN(regexp.MustCompile(`-gateway$`), GatewayHandler)

	mux.Use(line.EarlyTimeout(5000))   //gives handlers 5sec to clean up
	mux.Use(line.WithRuntimeSession()) //adds an aws session to the context
	mux.Use(line.ResourceAttributes()) //adds Terraform resource attributes

	return mux.Handle(msg, invoc)
}
