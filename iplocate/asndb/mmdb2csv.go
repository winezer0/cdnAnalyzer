package asndb

import (
	"encoding/csv"
	"fmt"
	"os"
)

// ExportASNToCSV 将数据库中的所有 ASN 条目导出为 CSV 文件
func ExportASNToCSV(outputPath string) error {
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

	// 遍历 IPv4 和 IPv6 数据库
	connectionIds := []string{"ipv4", "ipv6"}
	for _, connectionId := range connectionIds {
		reader, ok := mmDb[connectionId]
		if !ok {
			continue // 跳过未加载的数据库
		}

		networks := reader.Networks()
		for networks.Next() {
			var record ASNRecord
			ipNet, err := networks.Network(&record)
			if err != nil {
				fmt.Fprintf(os.Stderr, "解析网络段失败: %v\n", err)
				continue
			}

			row := []string{
				ipNet.String(),
				fmt.Sprintf("%d", record.AutonomousSystemNumber),
				record.AutonomousSystemOrg,
			}
			if err := writer.Write(row); err != nil {
				return fmt.Errorf("写入数据行失败: %v", err)
			}
		}

		if err := networks.Err(); err != nil {
			return fmt.Errorf("遍历数据库时发生错误: %v", err)
		}
	}

	fmt.Printf("成功导出 ASN 数据到: %s\n", outputPath)
	return nil
}
