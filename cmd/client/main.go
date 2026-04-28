package main

import (
	gohttp "github.com/repunit11/go-http"
)

func runClient() error {
	return gohttp.RunHTTP11PipelineClient(10)
}

func main() {
	if err := runClient(); err != nil {
		panic(err)
	}
}
