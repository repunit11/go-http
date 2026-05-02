package gohttp

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

const testServerModeEnv = "GOHTTP_TEST_SERVER_MODE"

func TestClientServer(t *testing.T) {
	cases := []struct {
		name   string
		server string
		client func(int) error
		count  int
	}{
		{name: "http10_basic", server: "basic", client: RunHTTP10BasicClient, count: 1000},
		{name: "http11_keep_alive", server: "keep-alive", client: RunHTTP11KeepAliveClient, count: 1000},
		{name: "http11_gzip", server: "gzip", client: RunHTTP11GzipClient, count: 1000},
		{name: "http11_chunked", server: "chunked", client: RunHTTP11ChunkedClient, count: 1},
		{name: "http11_pipeline", server: "pipeline", client: RunHTTP11PipelineClient, count: 1000},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// テストバイナリ自身を子プロセスとして再起動し、
			// TestHelperProcess に対象サーバーだけを起動させる。
			cmd := exec.Command(os.Args[0], "-test.run=^TestHelperProcess$")
			cmd.Env = append(os.Environ(), testServerModeEnv+"="+tc.server)

			serverLog := &bytes.Buffer{}
			cmd.Stdout = serverLog
			cmd.Stderr = serverLog

			if err := cmd.Start(); err != nil {
				t.Fatalf("failed to start server: %v", err)
			}
			t.Cleanup(func() {
				if cmd.Process != nil {
					_ = cmd.Process.Kill()
				}
				_ = cmd.Wait()
			})

			// サーバーのlisten開始を少し待ってから疎通確認する。
			time.Sleep(100 * time.Millisecond)

			// まずは素のTCP接続と1回のHTTPリクエストで、
			// サーバーが実際に応答できる状態か確認する。
			conn, err := net.DialTimeout("tcp", "localhost:8888", time.Second)
			if err != nil {
				t.Fatalf("server is not ready: %v\nserver log:\n%s", err, serverLog.String())
			}

			request, err := http.NewRequest("GET", "http://localhost:8888", nil)
			if err != nil {
				_ = conn.Close()
				t.Fatalf("failed to build probe request: %v", err)
			}
			if err := request.Write(conn); err != nil {
				_ = conn.Close()
				t.Fatalf("failed to send probe request: %v\nserver log:\n%s", err, serverLog.String())
			}

			response, err := http.ReadResponse(bufio.NewReader(conn), request)
			if err != nil {
				_ = conn.Close()
				t.Fatalf("server is not ready: %v\nserver log:\n%s", err, serverLog.String())
			}
			if tc.server != "chunked" {
				_, _ = io.Copy(io.Discard, response.Body)
			}
			_ = response.Body.Close()
			_ = conn.Close()

			// 起動確認が取れたら、本命のクライアント実装を実行して
			// 各HTTP挙動のテストケースを測定する。
			start := time.Now()
			if err := tc.client(tc.count); err != nil {
				t.Fatalf("client failed: %v\nserver log:\n%s", err, serverLog.String())
			}
			t.Logf("elapsed=%s", time.Since(start))
		})
	}
}

func TestHelperProcess(t *testing.T) {
	// 親テストから環境変数付きで起動されたときだけ、
	// 指定されたモードのサーバーをこの子プロセス内で動かす。
	switch os.Getenv(testServerModeEnv) {
	case "":
		return
	case "basic":
		RunHTTP10BasicServer()
	case "keep-alive":
		RunHTTP11KeepAliveServer()
	case "gzip":
		RunHTTP11GzipServer()
	case "chunked":
		RunHTTP11ChunkedServer()
	case "pipeline":
		RunHTTP11PipelineServer()
	default:
		t.Fatalf("unknown server mode: %q", os.Getenv(testServerModeEnv))
	}
}
