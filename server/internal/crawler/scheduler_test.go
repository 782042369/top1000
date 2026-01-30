package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"top1000/internal/model"
)

func TestFetchTop1000WithContext(t *testing.T) {
	t.Run("超时测试", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		_, err := FetchTop1000WithContext(ctx)
		if err == nil {
			t.Error("期望超时错误，但得到了 nil")
		}
	})
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name      string
		rawData   string
		wantCount int
		wantTime  string
		check     func([]model.SiteItem) error
	}{
		{
			name: "标准格式",
			rawData: `create time 2026-01-19 07:50:56 by xxx

站名：测试站点 【ID：123】
重复度：85.5%
文件大小：1.2TB
`,
			wantCount: 1,
			wantTime:  "2026-01-19 07:50:56",
			check: func(items []model.SiteItem) error {
				if items[0].SiteName != "测试站点" {
					t.Errorf("SiteName = %v, want %v", items[0].SiteName, "测试站点")
				}
				if items[0].SiteID != "123" {
					t.Errorf("SiteID = %v, want %v", items[0].SiteID, "123")
				}
				if items[0].Duplication != "85.5%" {
					t.Errorf("Duplication = %v, want %v", items[0].Duplication, "85.5%")
				}
				if items[0].Size != "1.2TB" {
					t.Errorf("Size = %v, want %v", items[0].Size, "1.2TB")
				}
				return nil
			},
		},
		{
			name: "多条数据",
			rawData: `create time 2026-01-19 07:50:56 by xxx

站名：站点1 【ID：1】
重复度：80%
文件大小：1TB
站名：站点2 【ID：2】
重复度：90%
文件大小：2TB
站名：站点3 【ID：3】
重复度：95%
文件大小：3TB
`,
			wantCount: 3,
			wantTime:  "2026-01-19 07:50:56",
		},
		{
			name: "Windows换行符",
			rawData: "create time 2026-01-19 07:50:56 by xxx\r\n" +
				"\r\n" +
				"站名：测试站点 【ID：123】\r\n" +
				"重复度：85.5%\r\n" +
				"文件大小：1.2TB\r\n",
			wantCount: 1,
			wantTime:  "2026-01-19 07:50:56",
		},
		{
			name: "数据行不完整(跳过)",
			rawData: `create time 2026-01-19 07:50:56 by xxx

站名：站点1 【ID：1】
重复度：80%
文件大小：1TB
站名：站点2 【ID：2】
重复度：90%
站名：站点3 【ID：3】
文件大小：3TB
`,
			wantCount: 2,
			wantTime:  "2026-01-19 07:50:56",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processed := parseResponse(tt.rawData)
			if processed.Time != tt.wantTime {
				t.Errorf("parseResponse() Time = %v, want %v", processed.Time, tt.wantTime)
			}
			if len(processed.Items) != tt.wantCount {
				t.Errorf("parseResponse() Items length = %v, want %v", len(processed.Items), tt.wantCount)
			}
			if tt.check != nil && len(processed.Items) > 0 {
				if err := tt.check(processed.Items); err != nil {
					t.Errorf("parseResponse() check failed: %v", err)
				}
			}
		})
	}
}

func TestExtractFieldValue(t *testing.T) {
	tests := []struct {
		name  string
		line  string
		want  string
	}{
		{
			name: "标准格式",
			line: "重复度：85.5%",
			want: "85.5%",
		},
		{
			name: "带空格",
			line: "文件大小： 1.2TB",
			want: "1.2TB",
		},
		{
			name: "无冒号",
			line: "纯文本",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractFieldValue(tt.line); got != tt.want {
				t.Errorf("extractFieldValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractTime(t *testing.T) {
	tests := []struct {
		name  string
		raw   string
		want  string
	}{
		{
			name: "标准格式",
			raw:  "create time 2026-01-19 07:50:56 by xxx",
			want: "2026-01-19 07:50:56",
		},
		{
			name: "无后缀",
			raw:  "create time 2026-01-19 07:50:56",
			want: "2026-01-19 07:50:56",
		},
		{
			name:  "空字符串",
			raw:   "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractTime(tt.raw); got != tt.want {
				t.Errorf("extractTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeLineEndings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Windows换行符",
			input: "line1\r\nline2\r\n",
			want:  "line1\nline2\n",
		},
		{
			name:  "Unix换行符",
			input: "line1\nline2\n",
			want:  "line1\nline2\n",
		},
		{
			name:  "混合换行符",
			input: "line1\r\nline2\nline3\r\n",
			want:  "line1\nline2\nline3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeLineEndings(tt.input); got != tt.want {
				t.Errorf("normalizeLineEndings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskMutex(t *testing.T) {
	t.Run("防止并发执行", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`create time 2026-01-19 07:50:56 by xxx

站名：测试站点 【ID：123】
重复度：85.5%
文件大小：1.2TB
`))
		}))
		defer server.Close()

		ctx := context.Background()
		errChan := make(chan error, 2)

		go func() {
			_, err := FetchTop1000WithContext(ctx)
			errChan <- err
		}()

		go func() {
			_, err := FetchTop1000WithContext(ctx)
			errChan <- err
		}()

		err1 := <-errChan
		err2 := <-errChan

		successCount := 0
		if err1 == nil {
			successCount++
		}
		if err2 == nil {
			successCount++
		}

		if successCount != 1 {
			t.Errorf("期望只有一个任务成功，实际有 %d 个成功", successCount)
		}
	})
}

func TestContextTimeout(t *testing.T) {
	t.Run("超时取消", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := FetchTop1000WithContext(ctx)
		if err == nil {
			t.Error("期望超时错误，但得到了 nil")
		}
	})
}
