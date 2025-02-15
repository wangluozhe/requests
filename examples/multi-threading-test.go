package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

func RequestServer(ctx context.Context, addr string) *http.Server {
	if addr == "" {
		addr = ":3334"
	}
	http.HandleFunc("/1k", func(w http.ResponseWriter, r *http.Request) {
		randomBytes := make([]byte, 1024*1)
		io.ReadFull(rand.Reader, randomBytes)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write(randomBytes)
	})

	http.HandleFunc("/10k", func(w http.ResponseWriter, r *http.Request) {
		randomBytes := make([]byte, 1024*10)
		io.ReadFull(rand.Reader, randomBytes)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write(randomBytes)
	})

	http.HandleFunc("/100k", func(w http.ResponseWriter, r *http.Request) {
		randomBytes := make([]byte, 1024*100)
		io.ReadFull(rand.Reader, randomBytes)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write(randomBytes)
	})
	Server := &http.Server{
		Addr: addr,
	}
	go Server.ListenAndServe()
	log.Print("准备")
	time.Sleep(time.Second * 4)
	log.Print("准备 ok")
	return Server
}
func RequestServer2(ctx context.Context, addr string) *http.Server {
	if addr == "" {
		addr = ":3334"
	}
	http.HandleFunc("/1k", func(w http.ResponseWriter, r *http.Request) {
		randomBytes := make([]byte, 1024*1)
		io.ReadFull(rand.Reader, randomBytes)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write(randomBytes)
	})

	http.HandleFunc("/10k", func(w http.ResponseWriter, r *http.Request) {
		randomBytes := make([]byte, 1024*10)
		io.ReadFull(rand.Reader, randomBytes)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write(randomBytes)
	})

	http.HandleFunc("/100k", func(w http.ResponseWriter, r *http.Request) {
		randomBytes := make([]byte, 1024*100)
		io.ReadFull(rand.Reader, randomBytes)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write(randomBytes)
	})
	Server := &http.Server{
		Addr: addr,
	}

	go Server.ListenAndServe()
	log.Print("准备")
	time.Sleep(time.Second * 4)
	log.Print("准备 ok")
	return Server
}

func main() {
	RequestServer(context.Background(), "127.0.0.1:3334")
	const (
		numRequests  = 10000 // 设置并发请求的数量
		urlToRequest = "http://127.0.0.1:3334/1k"
	)

	var wg sync.WaitGroup
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func(i int) {
			defer wg.Done()
			req := url.NewRequest()
			headers := url.NewHeaders()
			headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36") // 为了区分每个请求，稍微修改User-Agent
			req.Headers = headers
			r, err := requests.Get(urlToRequest, req)
			if err != nil {
				fmt.Printf("Request %d failed: %v\n", i, err)
				return
			}
			// 检查HTTP响应状态码
			if r.StatusCode != http.StatusOK {
				fmt.Printf("Request %d got status code: %d\n", i, r.StatusCode)
				return
			}
			fmt.Printf("Request %d response: %s\n", i, r.Text[:100]) // 只打印前100个字符以避免输出过多
		}(i)
	}

	wg.Wait() // 等待所有goroutine完成
}
