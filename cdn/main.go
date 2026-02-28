package main

import (
	"io"
	"log"
	"net/http"
)

// OriginServerの場所
const originURL = "http://localhost:8080"

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("proxy received:", r.Method, r.URL.Path)

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

	// OriginServerからのレスポンスをそのままクライアントに返す(ただの中継)
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Println("failed to copy response:", err)
	}
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