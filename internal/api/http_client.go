package api

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"
	"top1000/internal/config"
)

// getHTTPClient 从 context 提取超时时间，动态创建 HTTP 客户端
// 如果 context 有截止时间，使用它；否则使用默认超时
// 支持通过 INSECURE_SKIP_VERIFY 跳过 TLS 证书验证（用于证书过期等异常情况）
func getHTTPClient(ctx context.Context) *http.Client {
	// 如果 context 有截止时间，使用它
	timeout := defaultHTTPClientTimeout
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
		// 确保超时时间合理（大于0且小于默认超时）
		if timeout > 0 && timeout < defaultHTTPClientTimeout {
			// 使用 context 的超时
		} else {
			timeout = defaultHTTPClientTimeout
		}
	}

	// 检查是否需要跳过证书验证
	cfg := config.Get()
	client := &http.Client{Timeout: timeout}
	if cfg.InsecureSkipVerify {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = transport
	}

	return client
}
