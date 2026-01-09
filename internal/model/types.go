package model

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SiteItem 一条站点数据，Top1000里的每一项
type SiteItem struct {
	SiteName    string `json:"siteName"` // 站点名字，比如"朋友"、"馒头"
	SiteID      string `json:"siteid"`   // 站点资源ID，数字字符串
	Duplication string `json:"duplication"` // 重复度，比如"95%"
	Size        string `json:"size"`     // 文件大小，比如"1.5GB"
	ID          int    `json:"id"`       // 序号，从1开始，方便排序显示
}

// Validate 验证单条数据正确性
func (s *SiteItem) Validate() error {
	// 站点名称不能为空
	if strings.TrimSpace(s.SiteName) == "" {
		return fmt.Errorf("站点名称不能为空")
	}

	// 站点ID必须有，而且必须是数字
	if s.SiteID == "" {
		return fmt.Errorf("站点ID不能为空")
	}
	if _, err := strconv.ParseInt(s.SiteID, 10, 64); err != nil {
		return fmt.Errorf("站点ID必须是数字: %s", s.SiteID)
	}

	// 重复度可以空，但如果有值就得带%
	if s.Duplication != "" {
		if !strings.HasSuffix(s.Duplication, "%") {
			return fmt.Errorf("重复度格式错误: %s", s.Duplication)
		}
	}

	// 文件大小可以空，但如果有值就得符合格式
	if s.Size != "" {
		sizePattern := regexp.MustCompile(`^\d+(\.\d+)?\s*(KB|MB|GB|TB)$`)
		if !sizePattern.MatchString(s.Size) {
			return fmt.Errorf("文件大小格式错误: %s", s.Size)
		}
	}

	// ID必须大于0
	if s.ID <= 0 {
		return fmt.Errorf("ID必须大于0: %d", s.ID)
	}

	return nil
}

// ProcessedData 完整的Top1000数据，包含时间和1000条站点数据
type ProcessedData struct {
	Time  string     `json:"time"` // 数据时间，比如"2025-12-11 07:52:33"
	Items []SiteItem `json:"items"` // 站点数据列表，理论上应该有1000条
}

// Validate 验证完整数据，确保数据有效性
func (p *ProcessedData) Validate() error {
	// 时间不能为空
	if strings.TrimSpace(p.Time) == "" {
		return fmt.Errorf("时间不能为空")
	}

	// 至少需要有一条数据
	if len(p.Items) == 0 {
		return fmt.Errorf("数据条目不能为空")
	}

	// 逐条验证
	for i, item := range p.Items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("第%d条数据验证失败: %w", i+1, err)
		}
	}

	return nil
}
