package model

// SiteItem 表示top1000列表中的站点条目
type SiteItem struct {
	SiteName    string `json:"siteName"`
	SiteID      string `json:"siteid"`
	Duplication string `json:"duplication"`
	Size        string `json:"size"`
	ID          int    `json:"id"`
}

// ProcessedData 表示结构化数据
type ProcessedData struct {
	Time  string     `json:"time"`
	Items []SiteItem `json:"items"`
}
