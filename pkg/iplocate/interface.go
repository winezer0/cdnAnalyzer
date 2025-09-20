package iplocate

// IPInfo 接口定义了IP地址信息查询的标准方法
type IPInfo interface {
	// Find 查询单个IP地址的地理位置信息
	Find(query string) string

	// BatchFind 批量查询多个IP地址
	BatchFind(queries []string) map[string]string

	// GetDatabaseInfo 获取数据库信息
	GetDatabaseInfo() map[string]interface{}

	// Close 关闭数据库连接（清理资源）
	Close()
}

// IPv4Info 接口定义了IPv4地址信息查询的特定方法
type IPv4Info interface {
	IPInfo
}

// IPv6Info 接口定义了IPv6地址信息查询的特定方法
type IPv6Info interface {
	IPInfo
}

// IP2RegionInfo 接口定义了IP2Region查询的特定方法
type IP2RegionInfo interface {
	IPInfo
}
