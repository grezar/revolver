package main

import (
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/grezar/revolver"
	"github.com/grezar/revolver/reporting"
)

// These variables are set in build step.
var (
	Version  string
	Revision string
)

type MyRequest struct {
	Config string `json:"config"`
	DryRun bool   `json:"dry-run"`
}

func HandleRequest(req MyRequest) error {
	fmt.Println("revolver", Version, Revision)
	runner, err := revolver.NewRunner(req.Config, req.DryRun)
	if err != nil {
		return err
	}
	ok := reporting.Run(func(rptr *reporting.R) {
		runner.Run(rptr)
	})
	if !ok {
		return errors.New("failed to execute rotations")
	}
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
