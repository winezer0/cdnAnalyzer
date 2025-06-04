package asndb

import (
	"net"
	"sync"
)

// 全局缓存：ASN -> []*net.IPNet
var asnToIPNets map[uint64][]*net.IPNet
var cacheOnce sync.Once // 确保只初始化一次

// PreloadASNCache 预加载所有数据库条目到内存缓存中
func PreloadASNCache() error {
	cacheOnce.Do(func() {
		asnToIPNets = make(map[uint64][]*net.IPNet)
		connectionIds := []string{"ipv4", "ipv6"}

		for _, connectionId := range connectionIds {
			reader, ok := mmDb[connectionId]
			if !ok {
				continue
			}

			networks := reader.Networks()
			for networks.Next() {
				var record ASNRecord
				ipNet, err := networks.Network(&record)
				if err != nil {
					continue
				}

				asn := record.AutonomousSystemNumber
				asnToIPNets[asn] = append(asnToIPNets[asn], ipNet)
			}
		}
	})

	return nil
}

// FirstASNToIPRanges 快速查找：通过 ASN 查找对应的 IP 段
func FirstASNToIPRanges(targetASN uint64) ([]*net.IPNet, bool) {
	ipNets, found := asnToIPNets[targetASN]
	return ipNets, found
}
