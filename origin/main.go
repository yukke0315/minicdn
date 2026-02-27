package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    log.Println("request received:", r.Method, r.URL.Path)
	// クライアントに出力されるもの
	fmt.Fprintln(w, "Hello by Origin Server")
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Origin server is running on http://localhost:8080")
	// 監視するポートの指定
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}