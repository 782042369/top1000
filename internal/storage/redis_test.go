package storage

import (
	"context"
	"testing"
	"time"
	"top1000/internal/config"
	"top1000/internal/model"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestInitRedis(t *testing.T) {
	t.Run("成功连接Miniredis", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		redisClient = redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		if redisClient == nil {
			t.Error("redisClient 未初始化")
		}

		CloseRedis()
	})
}

func TestSaveDataWithContext(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	setupTestRedis(t, mr)
	defer CloseRedis()

	tests := []struct {
		name    string
		data    model.ProcessedData
		wantErr bool
	}{
		{
			name: "保存有效数据",
			data: model.ProcessedData{
				Time: "2026-01-19 07:50:56",
				Items: []model.SiteItem{
					{
						SiteName:    "测试站点",
						SiteID:      "123",
						Duplication: "85.5",
						Size:        "1.2TB",
						ID:          1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "保存多条数据",
			data: model.ProcessedData{
				Time: "2026-01-19 07:50:56",
				Items: []model.SiteItem{
					{SiteName: "站点1", SiteID: "1", ID: 1},
					{SiteName: "站点2", SiteID: "2", ID: 2},
				},
			},
			wantErr: false,
		},
		{
			name: "无效数据(验证失败)",
			data: model.ProcessedData{
				Time: "",
				Items: []model.SiteItem{
					{SiteName: "测试", SiteID: "1", ID: 1},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := SaveDataWithContext(ctx, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveDataWithContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadDataWithContext(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	setupTestRedis(t, mr)
	defer CloseRedis()

	ctx := context.Background()

	testData := model.ProcessedData{
		Time: "2026-01-19 07:50:56",
		Items: []model.SiteItem{
			{SiteName: "测试站点", SiteID: "123", Duplication: "85.5", Size: "1.2TB", ID: 1},
		},
	}

	_ = SaveDataWithContext(ctx, testData)

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
		check   func(*model.ProcessedData) error
	}{
		{
			name:    "成功加载数据",
			setup:   func() {},
			wantErr: false,
			check: func(data *model.ProcessedData) error {
				if data.Time != testData.Time {
					t.Errorf("Time = %v, want %v", data.Time, testData.Time)
				}
				if len(data.Items) != len(testData.Items) {
					t.Errorf("Items length = %v, want %v", len(data.Items), len(testData.Items))
				}
				return nil
			},
		},
		{
			name: "数据不存在",
			setup: func() {
				client := redisClient
				client.Del(ctx, config.DefaultRedisKey)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			data, err := LoadDataWithContext(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadDataWithContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				if err := tt.check(data); err != nil {
					t.Errorf("LoadDataWithContext() check failed: %v", err)
				}
			}
		})
	}
}

func TestDataExistsWithContext(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	setupTestRedis(t, mr)
	defer CloseRedis()

	ctx := context.Background()

	testData := model.ProcessedData{
		Time:  "2026-01-19 07:50:56",
		Items: []model.SiteItem{{SiteName: "测试", SiteID: "1", ID: 1}},
	}

	t.Run("数据存在", func(t *testing.T) {
		_ = SaveDataWithContext(ctx, testData)
		exists, err := DataExistsWithContext(ctx)
		if err != nil {
			t.Errorf("DataExistsWithContext() error = %v", err)
		}
		if !exists {
			t.Error("DataExistsWithContext() = false, want true")
		}
	})

	t.Run("数据不存在", func(t *testing.T) {
		redisClient.Del(ctx, config.DefaultRedisKey)
		exists, err := DataExistsWithContext(ctx)
		if err != nil {
			t.Errorf("DataExistsWithContext() error = %v", err)
		}
		if exists {
			t.Error("DataExistsWithContext() = true, want false")
		}
	})
}

func TestIsDataExpiredWithContext(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	setupTestRedis(t, mr)
	defer CloseRedis()

	ctx := context.Background()

	t.Run("数据过期", func(t *testing.T) {
		oldData := model.ProcessedData{
			Time:  "2020-01-01 00:00:00",
			Items: []model.SiteItem{{SiteName: "测试", SiteID: "1", ID: 1}},
		}
		_ = SaveDataWithContext(ctx, oldData)

		isExpired, err := IsDataExpiredWithContext(ctx)
		if err != nil {
			t.Errorf("IsDataExpiredWithContext() error = %v", err)
		}
		if !isExpired {
			t.Error("IsDataExpiredWithContext() = false, want true (数据应该过期)")
		}
	})

	t.Run("数据新鲜", func(t *testing.T) {
		freshData := model.ProcessedData{
			Time:  time.Now().Format("2006-01-02 15:04:05"),
			Items: []model.SiteItem{{SiteName: "测试", SiteID: "1", ID: 1}},
		}
		_ = SaveDataWithContext(ctx, freshData)

		isExpired, err := IsDataExpiredWithContext(ctx)
		if err != nil {
			t.Errorf("IsDataExpiredWithContext() error = %v", err)
		}
		if isExpired {
			t.Error("IsDataExpiredWithContext() = true, want false (数据应该新鲜)")
		}
	})

	t.Run("数据不存在", func(t *testing.T) {
		redisClient.Del(ctx, config.DefaultRedisKey)
		isExpired, err := IsDataExpiredWithContext(ctx)
		if err != nil {
			t.Errorf("IsDataExpiredWithContext() error = %v", err)
		}
		if !isExpired {
			t.Error("IsDataExpiredWithContext() = false, want true (数据不存在时应认为过期)")
		}
	})
}

func TestSaveSitesDataWithContext(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	setupTestRedis(t, mr)
	defer CloseRedis()

	ctx := context.Background()

	testData := map[string]interface{}{
		"site1": map[string]string{"name": "站点1"},
		"site2": map[string]string{"name": "站点2"},
	}

	err := SaveSitesDataWithContext(ctx, testData)
	if err != nil {
		t.Errorf("SaveSitesDataWithContext() error = %v", err)
	}

	exists, err := SitesDataExistsWithContext(ctx)
	if err != nil {
		t.Errorf("SitesDataExistsWithContext() error = %v", err)
	}
	if !exists {
		t.Error("SitesDataExistsWithContext() = false, want true")
	}

	loadedData, err := LoadSitesDataWithContext(ctx)
	if err != nil {
		t.Errorf("LoadSitesDataWithContext() error = %v", err)
	}

	loadedMap, ok := loadedData.(map[string]interface{})
	if !ok {
		t.Fatal("LoadSitesDataWithContext() 返回类型错误")
	}

	if len(loadedMap) != 2 {
		t.Errorf("LoadSitesDataWithContext() 返回 %d 条数据，期望 2 条", len(loadedMap))
	}
}

func TestIsUpdating(t *testing.T) {
	t.Run("默认不更新", func(t *testing.T) {
		if IsUpdating() {
			t.Error("IsUpdating() = true, want false")
		}
	})

	t.Run("设置更新标记", func(t *testing.T) {
		SetUpdating(true)
		if !IsUpdating() {
			t.Error("IsUpdating() = false, want true")
		}
		SetUpdating(false)
	})
}

func TestIsSitesUpdating(t *testing.T) {
	t.Run("默认不更新", func(t *testing.T) {
		if IsSitesUpdating() {
			t.Error("IsSitesUpdating() = true, want false")
		}
	})

	t.Run("设置更新标记", func(t *testing.T) {
		SetSitesUpdating(true)
		if !IsSitesUpdating() {
			t.Error("IsSitesUpdating() = false, want true")
		}
		SetSitesUpdating(false)
	})
}

func TestPing(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	setupTestRedis(t, mr)
	defer CloseRedis()

	t.Run("成功Ping", func(t *testing.T) {
		err := Ping()
		if err != nil {
			t.Errorf("Ping() error = %v", err)
		}
	})

	t.Run("Redis未初始化", func(t *testing.T) {
		oldClient := redisClient
		redisClient = nil
		defer func() { redisClient = oldClient }()

		err := Ping()
		if err == nil {
			t.Error("Ping() error = nil, want error")
		}
	})
}

func setupTestRedis(t *testing.T, mr *miniredis.Miniredis) {
	t.Helper()

	redisClient = redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
}

func TestSaveLoadDataRoundTrip(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	setupTestRedis(t, mr)
	defer CloseRedis()

	ctx := context.Background()

	original := model.ProcessedData{
		Time: "2026-01-19 07:50:56",
		Items: []model.SiteItem{
			{
				SiteName:    "测试站点",
				SiteID:      "123",
				Duplication: "85.5",
				Size:        "1.2TB",
				ID:          1,
			},
		},
	}

	err := SaveDataWithContext(ctx, original)
	if err != nil {
		t.Fatalf("SaveDataWithContext() error = %v", err)
	}

	loaded, err := LoadDataWithContext(ctx)
	if err != nil {
		t.Fatalf("LoadDataWithContext() error = %v", err)
	}

	if loaded.Time != original.Time {
		t.Errorf("Time = %v, want %v", loaded.Time, original.Time)
	}

	if len(loaded.Items) != len(original.Items) {
		t.Fatalf("Items length = %v, want %v", len(loaded.Items), len(original.Items))
	}

	for i := range original.Items {
		if loaded.Items[i].SiteName != original.Items[i].SiteName {
			t.Errorf("Items[%d].SiteName = %v, want %v", i, loaded.Items[i].SiteName, original.Items[i].SiteName)
		}
		if loaded.Items[i].SiteID != original.Items[i].SiteID {
			t.Errorf("Items[%d].SiteID = %v, want %v", i, loaded.Items[i].SiteID, original.Items[i].SiteID)
		}
	}
}
