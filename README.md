# solar-proxy





# デプロイ方法

## 1. LXCコンテナを用意

Ubunut 26で確認ずみ

下記のパッケージを導入しておく
```bash
apt install -y golang git
```


## 2. ソースのダウンロード

```bash
cd /opt
git clone https://github.com/gorakuru/solar-proxy.git
cd solar-proxy
```

## 3. ビルド

```bash
go mod init solar-proxy
go mod tidy
go build -o solar-proxy main.go
```

## 4. サービスファイルの用意



承知いたしました。LXCコンテナ内で直接Go環境を構築し、コードの作成からビルド、サービス化までをすべてコンテナ内で完結させる手順に変更します。

ホストPCでのクロスコンパイルやファイル転送の手間が省け、コンテナ単体で自己完結できるため管理がしやすくなります。ProxmoxのLXCコンソール（またはSSH）に `root` ユーザーでログインした状態から進めてください。

---

### 1. 必要なパッケージのインストール

まずはGo言語のコンパイラと、パッケージ取得に必要なGitをインストールします。

```bash
apt update
apt install -y golang git

```

### 2. プロジェクトの作成

今回は `/opt/solar-proxy` というディレクトリを作成し、そこで作業を行います。

```bash
mkdir -p /opt/solar-proxy
cd /opt/solar-proxy

# Goモジュールの初期化と必要パッケージのダウンロード
go mod init solar-proxy
go get github.com/icholy/digest

```

### 3. プログラムの作成

テキストエディタ（`nano`など）を使って `main.go` を作成します。

```bash
nano main.go

```

以下のコードをコピーして貼り付け、保存して閉じます。（`nano` の場合は `Ctrl+O` → `Enter` で保存、`Ctrl+X` で終了です）

```go
package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/icholy/digest"
)

func main() {
	// 環境変数の取得
	targetURL := os.Getenv("TARGET_URL")
	user := os.Getenv("DIGEST_USER")
	pass := os.Getenv("DIGEST_PASSWORD")
	port := os.Getenv("PORT")

	if targetURL == "" || user == "" || pass == "" {
		log.Fatal("エラー: TARGET_URL, DIGEST_USER, DIGEST_PASSWORD の環境変数が設定されていません。")
	}

	if port == "" {
		port = "80" // デフォルトはポート80
	}

	parsedTarget, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("TARGET_URLの解析エラー: %v", err)
	}

	// リバースプロキシの作成
	proxy := httputil.NewSingleHostReverseProxy(parsedTarget)

	// Digest認証を処理するトランスポートの設定
	proxy.Transport = &digest.Transport{
		Username: user,
		Password: pass,
	}

	// リクエストハンドラの設定
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// クライアントからのAuthorizationヘッダが残っていると競合するため削除
		r.Header.Del("Authorization")
		proxy.ServeHTTP(w, r)
	})

	log.Printf("リバースプロキシを起動しました: ポート %s -> 転送先 %s\n", port, targetURL)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

```

### 4. ビルドの実行

同じディレクトリ内でビルドコマンドを実行します。

```bash
go build -o solar-proxy main.go

```

エラーが出なければ、ディレクトリ内に `solar-proxy` という実行ファイルが作成されています。

### 5. systemdサービスの作成とデプロイ

作成した実行ファイルをバックグラウンドで動かし続けるために、systemdのサービスとして登録します。このファイル内で、リアルサーバーのIPや認証情報を環境変数として指定します。

```bash
nano /etc/systemd/system/solar-proxy.service

```

以下の内容を貼り付けて保存します。

```ini
[Unit]
Description=Solar Power Monitor Reverse Proxy (Digest Auth Bypass)
After=network.target

[Service]
Type=simple
# 実行ファイルのパスを /opt に変更しています
ExecStart=/opt/solar-proxy/solar-proxy
WorkingDirectory=/opt/solar-proxy
Restart=always
RestartSec=5

# === ここでデプロイ時の環境変数を設定 ===
Environment="TARGET_URL=http://192.168.68.120"
Environment="DIGEST_USER=user"
Environment="DIGEST_PASSWORD=12345678"
Environment="PORT=80"
# =========================================

User=root

[Install]
WantedBy=multi-user.target

```

### 6. サービスの起動と確認

systemdに変更を認識させ、サービスを起動します。

```bash
systemctl daemon-reload
systemctl enable solar-proxy
systemctl start solar-proxy

```

最後に、正常に起動しているかステータスを確認します。

```bash
systemctl status solar-proxy

```

緑色の文字で `active (running)` と表示され、ログに `リバースプロキシを起動しました: ポート 80 -> 転送先 http://192.168.68.120` と記録されていれば設定完了です。コンテナのIPアドレスへブラウザからアクセスし、認証なしで画面が表示されるか確認してください。
