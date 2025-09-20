package ip2region

import (
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"testing"
)

// TestIP2RegionIPv4 测试IPv4数据库查询
func TestIP2RegionIPv4(t *testing.T) {
	// xdb IPv4文件路径
	dbPath := `C:\Users\WINDOWS\Desktop\demo\ip2region_v4.xdb`

	// 创建IP2Region实例
	ip2Region, err := NewIP2Region(xdb.IPv4, dbPath)
	if err != nil {
		t.Fatalf("创建IPv4 IP2Region实例失败: %v", err)
	}
	defer ip2Region.Close()

	// 测试单个IP查询
	testIP := "8.8.8.8"
	result := ip2Region.Find(testIP)
	if result == "" {
		t.Errorf("查询IP %s 失败，返回空结果", testIP)
	} else {
		t.Logf("查询IP %s 结果: %s", testIP, result)
	}

	// 测试批量查询
	testIPs := []string{"114.114.114.114", "1.1.1.1", "8.8.4.4"}
	batchResults := ip2Region.BatchFind(testIPs)
	for ip, result := range batchResults {
		if result == "" {
			t.Errorf("批量查询IP %s 失败，返回空结果", ip)
		} else {
			t.Logf("批量查询IP %s 结果: %s", ip, result)
		}
	}

	// 测试获取数据库信息
	dbInfo := ip2Region.GetDatabaseInfo()
	if len(dbInfo) == 0 {
		t.Error("获取数据库信息失败")
	} else {
		t.Logf("数据库信息: %+v", dbInfo)
	}
}

// TestIP2RegionIPv6 测试IPv6数据库查询
func TestIP2RegionIPv6(t *testing.T) {
	// xdb IPv6文件路径
	dbPath := `C:\Users\WINDOWS\Desktop\demo\ip2region_v6.xdb`

	// 创建IP2Region实例
	ip2Region, err := NewIP2Region(xdb.IPv6, dbPath)
	if err != nil {
		t.Fatalf("创建IPv6 IP2Region实例失败: %v", err)
	}
	defer ip2Region.Close()

	// 测试单个IPv6查询
	testIP := "2001:4860:4860::8888"
	result := ip2Region.Find(testIP)
	if result == "" {
		t.Errorf("查询IPv6 %s 失败，返回空结果", testIP)
	} else {
		t.Logf("查询IPv6 %s 结果: %s", testIP, result)
	}

	// 测试批量IPv6查询
	testIPs := []string{"2001:4860:4860::8844", "2606:4700:4700::1111"}
	batchResults := ip2Region.BatchFind(testIPs)
	for ip, result := range batchResults {
		if result == "" {
			t.Errorf("批量查询IPv6 %s 失败，返回空结果", ip)
		} else {
			t.Logf("批量查询IPv6 %s 结果: %s", ip, result)
		}
	}

	// 测试获取数据库信息
	dbInfo := ip2Region.GetDatabaseInfo()
	if len(dbInfo) == 0 {
		t.Error("获取数据库信息失败")
	} else {
		t.Logf("数据库信息: %+v", dbInfo)
	}
}

// TestIP2RegionWithVectorIndex 测试使用 VectorIndex 缓存的查询
func TestIP2RegionWithVectorIndex(t *testing.T) {
	// xdb IPv4文件路径
	dbPath := `C:\Users\WINDOWS\Desktop\demo\ip2region_v4.xdb`

	// 预加载VectorIndex
	vectorIndex, err := LoadVectorIndexFromFile(dbPath)
	if err != nil {
		t.Fatalf("加载VectorIndex失败: %v", err)
	}

	// 创建使用VectorIndex的IP2Region实例
	ip2Region, err := NewIP2RegionWithVectorIndex(xdb.IPv4, dbPath, vectorIndex)
	if err != nil {
		t.Fatalf("创建使用VectorIndex的IP2Region实例失败: %v", err)
	}
	defer ip2Region.Close()

	// 测试查询
	testIP := "114.114.114.114"
	result := ip2Region.Find(testIP)
	if result == "" {
		t.Errorf("使用VectorIndex查询IP %s 失败，返回空结果", testIP)
	} else {
		t.Logf("使用VectorIndex查询IP %s 结果: %s", testIP, result)
	}
}

// TestIP2RegionVerify 测试xdb文件验证功能
func TestIP2RegionVerify(t *testing.T) {
	// 验证IPv4 xdb文件
	dbPath := `C:\Users\WINDOWS\Desktop\demo\ip2region_v4.xdb`
	err := xdb.VerifyFromFile(dbPath)
	if err != nil {
		t.Errorf("验证IPv4 xdb文件失败: %v", err)
	} else {
		t.Log("IPv4 xdb文件验证成功")
	}

	// 验证IPv6 xdb文件
	dbPath = `C:\Users\WINDOWS\Desktop\demo\ip2region_v6.xdb`
	err = xdb.VerifyFromFile(dbPath)
	if err != nil {
		t.Errorf("验证IPv6 xdb文件失败: %v", err)
	} else {
		t.Log("IPv6 xdb文件验证成功")
	}
}
