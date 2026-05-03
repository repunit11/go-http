package gohttp

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
)

type clientSession struct {
	conn   net.Conn
	reader *bufio.Reader
}
type clienthandler func(*clientSession, int) error

func RunHTTP10BasicClient(count int) error {
	return runClient(count, processBasic)
}

func RunHTTP11KeepAliveClient(count int) error {
	return runClient(count, processKeepAlive)
}

func RunHTTP11GzipClient(count int) error {
	return runClient(count, processGzip)
}

func RunHTTP11ChunkedClient(count int) error {
	return runClient(count, processChunked)
}

func RunHTTP11PipelineClient(count int) error {
	var conn net.Conn = nil
	var err error
	conn, err = net.Dial("tcp", "localhost:8888")
	if err != nil {
		return err
	}
	fmt.Printf("Access\n")
	defer conn.Close()

	requests := make([]*http.Request, 0, count)
	for i := 0; i < count; i++ {
		lastMessage := i == count-1
		request, err := http.NewRequest(
			"GET", "http://localhost:8888?message="+strconv.Itoa(i), nil)
		if err != nil {
			return err
		}
		if lastMessage {
			request.Header.Add("Connection", "close")
		} else {
			request.Header.Add("Connection", "keep-alive")
		}
		err = request.Write(conn)
		if err != nil {
			return err
		}
		fmt.Println("send: ", i)
		requests = append(requests, request)
	}

	reader := bufio.NewReader(conn)
	for _, request := range requests {
		// pipeline の学習ポイントとして、この順序付き受信は維持したい。
		// request を送った順に response を対応付けて読む。
		response, err := http.ReadResponse(reader, request)
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

func runClient(count int, handler clienthandler) error {
	session := &clientSession{}
	for i := 0; i < count; i++ {
		err := handler(session, i)
		if err != nil {
			return err
		}
	}

	return closeSession(session)
}

func connect(session *clientSession, current int) error {
	if session.conn != nil {
		return nil
	}

	conn, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		return err
	}
	session.conn = conn
	session.reader = bufio.NewReader(conn)
	fmt.Printf("Access: %d\n", current)
	return nil
}

func closeSession(session *clientSession) error {
	if session.conn == nil {
		session.reader = nil
		return nil
	}

	err := session.conn.Close()
	session.conn = nil
	session.reader = nil
	return err
}

func processChunked(session *clientSession, current int) error {
	// コネクションを張る
	if err := connect(session, current); err != nil {
		return err
	}
	// リクエストの作成、書き込み
	request, err := http.NewRequest("GET", "http://localhost:8888", nil)
	if err != nil {
		return err
	}
	err = request.Write(session.conn)
	if err != nil {
		return err
	}

	// レスポンスの受け取り
	response, err := http.ReadResponse(session.reader, request)
	if err != nil {
		return err
	}
	dump, err := httputil.DumpResponse(response, false)
	if err != nil {
		return err
	}
	fmt.Println(string(dump))

	// チャンクごとに受け取り
	if len(response.TransferEncoding) < 1 || response.TransferEncoding[0] != "chunked" {
		return fmt.Errorf("wrong transfer encoding")
	}
	for {
		sizeStr, err := session.reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		size, err := strconv.ParseInt(string(sizeStr[:len(sizeStr)-2]), 16, 64)
		if size == 0 {
			break
		}
		if err != nil {
			return err
		}

		line := make([]byte, int(size))
		io.ReadFull(session.reader, line)
		session.reader.Discard(2)
		fmt.Printf("	%d bytes: %s\n", size, string(line))
	}
	return nil
}

func processGzip(session *clientSession, current int) error {
	// コネクションを張る
	if err := connect(session, current); err != nil {
		return err
	}

	// リクエストの作成、書き込み
	request, err := http.NewRequest(
		"GET", "http://localhost:8888", nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept-Encoding", "gzip")

	err = request.Write(session.conn)
	if err != nil {
		return err
	}

	// レスポンスの受け取り
	response, err := http.ReadResponse(
		session.reader, request)
	if err != nil {
		fmt.Println("Retry")
		closeSession(session)
		return err
	}

	// 結果の表示
	dump, err := httputil.DumpResponse(response, false)
	if err != nil {
		return err
	}
	fmt.Println(string(dump))

	defer response.Body.Close()

	if response.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(response.Body)
		if err != nil {
			return err
		}
		io.Copy(os.Stdout, reader)
		reader.Close()
	} else {
		io.Copy(os.Stdout, response.Body)
	}

	return nil
}

func processKeepAlive(session *clientSession, current int) error {
	// コネクションを張る
	if err := connect(session, current); err != nil {
		return err
	}

	// リクエストの作成、書き込み
	request, err := http.NewRequest(
		"GET", "http://localhost:8888", nil)
	if err != nil {
		return err
	}
	err = request.Write(session.conn)
	if err != nil {
		return err
	}

	// レスポンスの受け取り
	response, err := http.ReadResponse(
		session.reader, request)
	if err != nil {
		fmt.Println("Retry")
		closeSession(session)
		return err
	}

	// 結果の表示
	dump, err := httputil.DumpResponse(response, true)
	if err != nil {
		return err
	}
	fmt.Println(string(dump))

	return nil
}

func processBasic(session *clientSession, current int) error {
	connValue, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		return err
	}
	defer connValue.Close()
	request, err := http.NewRequest(
		"GET", "http://localhost:8888", nil)
	if err != nil {
		return err
	}
	request.Write(connValue)
	response, err := http.ReadResponse(
		bufio.NewReader(connValue), request)
	if err != nil {
		return err
	}
	dump, err := httputil.DumpResponse(response, true)
	if err != nil {
		return err
	}
	fmt.Println(string(dump))
	return nil
}
