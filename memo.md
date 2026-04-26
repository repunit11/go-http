## Server
### 初期状態
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

### Keep-Alive対応
1. forでリクエストの受付をループしている
   TCPコネクションが張られた後に何度もリクエストを受け付ける
2. conn.SetReadDeadlineでタイムアウトを設定しておく
3. http.ReadRequestでリクエストを待つ
   タイムアウトしたらerrにnet.Errorをラップした構造体が設定

### 圧縮
1. クライアントがgzip圧縮を認めているかを判定する
2. gzip.NewWriterに本文を流し込んで圧縮する

### チャンク形式
1. contentsの各要素について、長さと本文を書き込む
2. 最後に終了チャンクを書く

### パイプライニング
1. チャネルを使うことで送受信を非同期化する
2. リクエストに対応するレスポンス受け取り口を順番キューに積む
3. リクエスト処理はgoroutineで進める
4. キュー順にレスポンス受け取り口を待つ

## Client
### 初期状態
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

### Keep-Alive対応
1. コネクションが切れていたら再接続をするようにしている
   コネクションをわざと切るようにsleepを追加

### チャンク形式
1. チャンクごとに受け取る
2. \r\nまでを読み取る
3. \r\nを除いた文字列の長さを16進数に変換する
4. lineというバッファに読み込んだ文字列を書き込んで\r\nを取り除く

### パイプライニング
1. バッファを作成してリクエストを詰める
2. それぞれのリクエストに合ったレスポンスを取得する 
ちなみにReadResponseにリクエストが引数として必要なのはレスポンス構造体にリクエストが含まれているから

## 計測

### HTTP/1.0
- 1.593s
- 1.571s
- 1.566s

### HTTP/1.1 Keep-Alive
- 0.387s
- 0.378s
- 0.393s

### HTTP/1.1 圧縮
- 0.815s
- 0.852s
- 0.819s

### HTTP/1.1 チャンク
- 0.400s
- 0.419s
- 0.459s

### HTTP/1.1 パイプ (+ Keep-Alive)
- 0.282s
- 0.270s
- 0.318s
