package gohttp

import (
	"testing"
	"time"
)

func TestRunClientDuration(t *testing.T) {
	go func() {
		RunServer11()
	}()
	time.Sleep(100 * time.Millisecond)

	start := time.Now()
	err := RunClient11Pipe(1000)
	elapsed := time.Since(start)

	t.Logf("elapsed=%s", elapsed)

	if err != nil {
		t.Fatal(err)
	}
}
