package main

import (
	"fmt"
	"log"
	"time"

	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/models"
	"github.com/wangluozhe/requests/url"
)

// 1. 定义一个日志中间件
// 它会在请求发送前打印 "Start..."，请求结束后打印 "End..." 和耗时
func LoggingMiddleware(next requests.Handler) requests.Handler {
	return func(preq *models.PrepareRequest, req *url.Request) (*models.Response, error) {
		start := time.Now()
		log.Printf("[Middleware] 开始请求: %s %s", preq.Method, preq.Url)

		// 调用下一个处理程序（或者是真正的发送逻辑）
		resp, err := next(preq, req)

		duration := time.Since(start)
		if err != nil {
			log.Printf("[Middleware] 请求失败: %v (耗时: %v)", err, duration)
		} else {
			log.Printf("[Middleware] 请求完成: 状态码 %d (耗时: %v)", resp.StatusCode, duration)
		}

		return resp, err
	}
}

// 2. 定义一个鉴权中间件
// 它会强制给所有经过该 Session 的请求加上 Authorization 头
func AuthMiddleware(token string) requests.Middleware {
	return func(next requests.Handler) requests.Handler {
		return func(preq *models.PrepareRequest, req *url.Request) (*models.Response, error) {
			// 在发送前修改请求头
			// 注意：这里修改 preq.Headers (底层的 http.Header)
			if preq.Headers != nil {
				preq.Headers.Set("Authorization", "Bearer "+token)
			} else {
				// 如果 header 为空，可能需要初始化（视具体逻辑而定，通常 prepare 阶段已有 header）
				// 这里简单打印一下演示
				log.Println("[Middleware] Injecting Auth Token...")
			}

			// 也可以修改 req *url.Request (用户层面的配置)
			// req.Headers 可能需要通过 req.GetHeaders() 获取并修改

			log.Printf("[Middleware] 已添加鉴权 Token: %s", token)

			// 继续执行
			return next(preq, req)
		}
	}
}

func main() {
	// 创建 Session
	s := requests.NewSession()

	// 3. 注册中间件
	// 注意顺序：先注册的先执行前置逻辑，后执行后置逻辑（洋葱模型）
	// 这里的执行顺序是：
	// Logging Start -> Auth Logic -> Send Request -> Auth End (no logic) -> Logging End
	s.Use(LoggingMiddleware)
	s.Use(AuthMiddleware("my-secret-token-123"))

	// 发送请求
	// 不需要每次都去定义 Hook，中间件会自动生效
	resp, err := s.Get("https://httpbin.org/get", nil)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("--------------------------------------------------")
	fmt.Printf("最终响应内容: %s\n", resp.Text)
	fmt.Println("--------------------------------------------------")

	// 验证 headers 是否包含 Authorization
	// httpbin 会返回请求头，我们可以检查一下
	fmt.Println("检查 httpbin 返回的 headers (确认中间件是否生效):")
	fmt.Println(resp.Text)
}
