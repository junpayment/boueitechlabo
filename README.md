# Go MITM Proxy Server

このプロジェクトは、Go言語を使用して実装されたTLSインスペクション（MITM）プロキシサーバーの学習用プロトタイプです。
HTTPS通信をインターセプトし、動的に生成された証明書を使用して暗号化通信の内容を解析・計測します。

## 1. 事前準備

### 認証局 (CA) の作成
通信を傍受するために、独自のルート証明書と秘密鍵を作成する必要があります。

```bash
# 秘密鍵の生成
openssl genrsa -out ca.key 2048

# 自作CA証明書の生成
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt \
    -subj "/CN=MITM Proxy CA/O=Bouei Tech Labo"
```

### 環境変数の設定
プロジェクトのルートにある `.envrc` またはシェルで以下の環境変数を設定してください。

```shell script
export CERT_PATH=$(pwd)/ca.crt
export KEY_PATH=$(pwd)/ca.key
```


## 2. 起動方法

依存関係を解決し、サーバーを起動します。

```shell script
# 依存パッケージのインストール
go mod tidy

# サーバーの起動 (ポート :8080 で待機)
go run main.go
```


## 3. 動作確認

別のターミナルから `curl` を使用して、プロキシ経由でHTTPS通信を行います。
この際、作成した `ca.crt` を信頼する証明書として指定する必要があります。

```shell script
# -x: プロキシを指定
# --cacert: 自作CA証明書を指定して検証をパスさせる
curl -v -x http://localhost:8080 --cacert ./ca.crt https://example.com
```


### 実行ログの確認
サーバー側のターミナルに以下のような通信ログが表示されれば成功です。

```plain text
[start] example.com:443 from 93.184.215.14
host: example.com, port: 443
C->S copy finished
S->C copy finished
[end] example.com:443 from 93.184.215.14 - sent: 74 bytes, received: 885 bytes
```

## 注意事項
- このツールは学習およびデバッグ目的でのみ使用してください。
- ブラウザでテストする場合は、`ca.crt` をシステムの「信頼されたルート証明機関」にインポートする必要があります。
- 実用的な利用には、証明書のキャッシュ機構やより詳細なHTTP解析ロジックの追加が推奨されます。
 
### ポイント
*   **CA作成コマンド**: MITMには必須のステップなので、そのままコピペで動くように記載しました。
*   **環境変数**: `main.go` で `os.Getenv` を使っているため、設定方法を明記しています。
*   **curlコマンド**: 先ほどのエラーを踏まえ、`--cacert ./ca.crt` を正しく使う方法を記載しています。
