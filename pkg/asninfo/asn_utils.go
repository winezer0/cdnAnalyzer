package asninfo

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/winezer0/cdnAnalyzer/pkg/logging"
)

// ExportToCSV 将数据库中的所有 ASN 条目导出为 CSV 文件
func (m *MMDBManager) ExportToCSV(outputPath string) error {
	// 创建输出文件
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 创建 CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	header := []string{"CIDR", "ASN", "Organization"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("写入表头失败: %v", err)
	}

	m.mu.RLock()
	reader := m.mmDb
	m.mu.RUnlock()

	if reader == nil {
		return fmt.Errorf("数据库未初始化")
	}

	networks := reader.Networks()
	for networks.Next() {
		var record ASNRecord
		ipNet, err := networks.Network(&record)
		if err != nil {
			logging.Errorf("解析网络段失败: %v", err)
			continue
		}

		row := []string{
			ipNet.String(),
			fmt.Sprintf("%d", record.AutonomousSystemNumber),
			record.AutonomousSystemOrganization,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("写入数据行失败: %v", err)
		}
	}

	if err := networks.Err(); err != nil {
		return fmt.Errorf("遍历数据库时发生错误: %v", err)
	}

	logging.Debugf("成功导出 ASN 数据到: %s", outputPath)
	return nil
}

// GetUniqueOrgNumbers 提取 FoundASN == true 的 OrganisationNumber，并去重
func GetUniqueOrgNumbers(asnInfos []ASNInfo) []uint64 {
	seen := make(map[uint64]bool)
	var result []uint64

	for _, info := range asnInfos {
		if info.FoundASN {
			orgNum := info.OrganisationNumber
			if !seen[orgNum] {
				seen[orgNum] = true
				result = append(result, orgNum)
			}
		}
	}

	return result
}

// PrintASNInfo 打印ASN信息
func PrintASNInfo(ipInfo *ASNInfo) {
	if ipInfo == nil {
		logging.Debug("ASN信息为空")
		return
	}

	logging.Debugf("IP: %15s | 版本: %d | 找到ASN: %v | ASN: %6d | 组织: %s",
		ipInfo.IP,
		ipInfo.IPVersion,
		ipInfo.FoundASN,
		ipInfo.OrganisationNumber,
		ipInfo.OrganisationName,
	)
}

// getIpVersion 获取IP版本号
func getIpVersion(ipString string) int {
	ipVersion := 4
	if strings.Contains(ipString, ":") {
		ipVersion = 6
	}

	return ipVersion
}
