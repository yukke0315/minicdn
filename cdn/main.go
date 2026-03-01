package main

import (
	"io"
	"log"
	"net/http"
)

// OriginServerの場所
const originURL = "http://localhost:8080"

// キャッシュ用。URL->レスポンス本文
var cache = map[string][]byte{}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("proxy received:", r.Method, r.URL.Path)

	key := r.URL.String()

	// キャッシュの確認
	if data, ok := cache[key]; ok {
		log.Println("cache hit:", key)
		w.Write(data)
		return
	} else {
		log.Println("cache miss:", key)
	}

	// クライアントからのリクエストをOriginServerに転送する
	targetURL := originURL + r.URL.Path
	
	// メソッド、ボディは保持してURLのみ変更してリクエスト作成
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header = r.Header

	// OriginServerに送信
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "failed to reach origin", http.StatusBadGateway)
		return
	}
	// 関数の最後で必ずbodyを閉じる。retunで抜けても動く
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	var body []byte
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read response", http.StatusInternalServerError)
		return
	}

	// キャッシュに保存
	cache[key] = body

	// クライアントに返す
	w.Write(body)
}

func main() {
	http.HandleFunc("/", proxyHandler)
	log.Println("miniCDN is running on http://localhost:8081")
	// 監視するポートの指定
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}