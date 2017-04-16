package main

import (
	"testing"

	"github.com/microfactory/line"
)

func TestHandle(t *testing.T) {

	res, err := Handle([]byte(`{}`), &line.Invocation{
		InvokedFunctionARN:    "test:arn-gateway",
		RemainingTimeInMillis: func() int64 { return 3000 },
	})
	if err != nil {
		t.Fatal(err)
	}

	_ = res

}
