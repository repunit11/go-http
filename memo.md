## Server
1. net.Listen
   TCP用のソケットを作る→localhost:8888に紐づける→接続待ち状態にする
2. listner.Accept
   SYNパケットを待つ→SYNパケットを受け取りコネクションが確立されるとconnを返す
3. goroutine
   Acceptで受け取ったconnの処理を渡す。各リクエストを処理する
4. conn
   バイト列を読み書きする
5. http.ReadRequest
   connに入っているバイト列をGoのhttp.Requestの構造体に変換する
6. httputil.DumpRequest
   構造体に変換したrequestをテキストに変換して表示。ログっぽい用途
7. response.Write
   HTTPレスポンスをTCPに流せるバイト列に変換して、connに書き込む
8. conn.Close
   TCPコネクションを閉じる

## Client
1. net.Dial
   localhost:8888で待ち受けているサーバーに対してTCP接続を開始する（SYNパケットを送る）
2. http.NewRequest
   HTTPリクエストを作る
3. request.Write
   requestがHTTPのテキスト形式に変換されてそれがconnに書き込まれる
4. http.ReadResponse
   connにレスポンスが書き込まれているのでバイト列をhttp.Response構造体に変換する
5. bufio.NewReader
   connから直接読むのではなくバッファ付きReaderで包んでいる。行単位、必要量単位で処理するため。
6. httputil.DumpResponse
   responseをログ表示させやすいHTTPテキスト風の形に戻している
