package api

import (
	"context"
	"net/http"
	"time"
)

// getHTTPClient 从 context 提取超时时间，动态创建 HTTP 客户端
// 如果 context 有截止时间，使用它；否则使用默认超时
func getHTTPClient(ctx context.Context) *http.Client {
	// 如果 context 有截止时间，使用它
	if deadline, ok := ctx.Deadline(); ok {
		timeout := time.Until(deadline)
		// 确保超时时间合理（大于0且小于默认超时）
		if timeout > 0 && timeout < defaultHTTPClientTimeout {
			return &http.Client{Timeout: timeout}
		}
	}

	// 使用默认超时
	return &http.Client{Timeout: defaultHTTPClientTimeout}
}
