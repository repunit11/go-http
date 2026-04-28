package gohttp

import (
	"testing"
	"time"
)

func TestRunClientDuration(t *testing.T) {
	go func() {
		RunHTTP11PipelineServer()
	}()
	time.Sleep(100 * time.Millisecond)

	start := time.Now()
	err := RunHTTP11PipelineClient(1000)
	elapsed := time.Since(start)

	t.Logf("elapsed=%s", elapsed)

	if err != nil {
		t.Fatal(err)
	}
}
