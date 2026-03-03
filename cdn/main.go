package main

import (
	"io"
	"log"
	"net/http"
	"time"
)

// OriginServerの場所
const originURL = "http://localhost:8080"

// キャッシュ用構造体
type CacheEntry struct {
	body []byte
	expiresAt time.Time
}

// キャッシュ用。URL->レスポンス本文
var cache = map[string]CacheEntry{}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("proxy received:", r.Method, r.URL.Path)

	key := r.URL.String()

	// キャッシュの確認
	if entry, ok := cache[key]; ok {
		// 有効期限の確認
		if time.Now().Before(entry.expiresAt) {
			log.Println("cache hit:", key)
			w.Write(entry.body)
			return
		} else {
			log.Println("cache expired:", key)
			delete(cache, key)
		}
	}

	log.Println("cache miss:", key)

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
	cache[key] = CacheEntry{
		body: body,
		// キャッシュの期限（60s）
		expiresAt: time.Now().Add(60 * time.Second),
	}

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