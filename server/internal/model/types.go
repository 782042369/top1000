package model

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var sizePattern = regexp.MustCompile(`^\d+(\.\d+)?\s*(KB|MB|GB|TB)$`)

// SiteItem 一条站点数据
type SiteItem struct {
	SiteName    string `json:"siteName"`
	SiteID      string `json:"siteid"`
	Duplication string `json:"duplication"`
	Size        string `json:"size"`
	ID          int    `json:"id"`
}

// Validate 验证单条数据正确性
func (s *SiteItem) Validate() error {
	if strings.TrimSpace(s.SiteName) == "" {
		return fmt.Errorf("站点名称不能为空")
	}

	if s.SiteID == "" {
		return fmt.Errorf("站点ID不能为空")
	}
	if _, err := strconv.ParseInt(s.SiteID, 10, 64); err != nil {
		return fmt.Errorf("站点ID必须是数字: %s", s.SiteID)
	}

	if s.Duplication != "" {
		if _, err := strconv.ParseFloat(s.Duplication, 64); err != nil {
			return fmt.Errorf("重复度必须为数字: %s", s.Duplication)
		}
	}

	if s.Size != "" && !sizePattern.MatchString(s.Size) {
		return fmt.Errorf("文件大小格式错误: %s", s.Size)
	}

	if s.ID <= 0 {
		return fmt.Errorf("ID必须大于0: %d", s.ID)
	}

	return nil
}

// ProcessedData 完整的Top1000数据
type ProcessedData struct {
	Time  string     `json:"time"`
	Items []SiteItem `json:"items"`
}

// Validate 验证完整数据
func (p *ProcessedData) Validate() error {
	if strings.TrimSpace(p.Time) == "" {
		return fmt.Errorf("时间不能为空")
	}

	if len(p.Items) == 0 {
		return fmt.Errorf("数据条目不能为空")
	}

	for i, item := range p.Items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("第%d条数据验证失败: %w", i+1, err)
		}
	}

	return nil
}
