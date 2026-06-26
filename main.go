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
