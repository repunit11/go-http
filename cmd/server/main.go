package main

import (
	gohttp "github.com/repunit11/go-http"
)

func runServer() {
	gohttp.RunHTTP11PipelineServer()
}

func main() {
	runServer()
}
