package main

import "encoding/json"

// Context provides information about Lambda execution environment.
type Context struct {
	FunctionName          string       `json:"function_name"`
	FunctionVersion       string       `json:"function_version"`
	InvokedFunctionARN    string       `json:"invoked_function_arn"`
	MemoryLimitInMB       int          `json:"memory_limit_in_mb,string"`
	AWSRequestID          string       `json:"aws_request_id"`
	LogGroupName          string       `json:"log_group_name"`
	LogStreamName         string       `json:"log_stream_name"`
	RemainingTimeInMillis func() int64 `json:"-"`
}

//Handle lambda events
func Handle(ev json.RawMessage, ctx *Context) (interface{}, error) {
	//@TODO the Line library should do:
	// - multiplex multiple lambda functions onto some handlers
	// - allow little boilerplate api gateway proxy handling
	// - add better typing of aws events from certain Streams
	// - automatically map and initialize aws services from a policy document
	// - make arbitrary environment variables available

	return "hello world", nil
}
