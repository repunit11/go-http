package gohttp

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
)

func RunClient11(count int) error {
	current := 0
	var conn net.Conn
	for i := 0; i < count; i++ {
		var err error
		// コネクションを張る
		if conn == nil {
			conn, err = net.Dial("tcp", "localhost:8888")
			if err != nil {
				return err
			}
			fmt.Printf("Access: %d\n", current)
		}

		// リクエストの作成、書き込み
		request, err := http.NewRequest(
			"GET", "http://localhost:8888", nil)
		if err != nil {
			return err
		}
		err = request.Write(conn)
		if err != nil {
			return err
		}

		// レスポンスの受け取り
		response, err := http.ReadResponse(
			bufio.NewReader(conn), request)
		if err != nil {
			fmt.Println("Retry")
			conn = nil
			continue
		}

		// 結果の表示
		dump, err := httputil.DumpResponse(response, true)
		if err != nil {
			return err
		}
		fmt.Println(string(dump))

		current++
	}

	if conn != nil {
		return conn.Close()
	}
	return nil
}

func RunClient10(count int) error {
	for i := 0; i < count; i++ {
		conn, err := net.Dial("tcp", "localhost:8888")
		if err != nil {
			return err
		}
		defer conn.Close()
		request, err := http.NewRequest(
			"GET", "http://localhost:8888", nil)
		if err != nil {
			return err
		}
		request.Write(conn)
		response, err := http.ReadResponse(
			bufio.NewReader(conn), request)
		if err != nil {
			return err
		}
		dump, err := httputil.DumpResponse(response, true)
		if err != nil {
			return err
		}
		fmt.Println(string(dump))
	}
	return nil
}
