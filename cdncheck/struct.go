package cdncheck

import (
	"fmt"
	"github.com/oschwald/geoip2-golang"
	"github.com/yl2chen/cidranger"
	"path/filepath"
)

type CDNCheck struct {
	ranger   cidranger.Ranger
	geoip2Db *geoip2.Reader
}

// NewCDNCheck 创建CDNCheck对象
func NewCDNCheck() *CDNCheck {
	cdn := CDNCheck{}
	_ = cdn.LoadGeoASNDb()
	return &cdn
}

func (c *CDNCheck) LoadGeoASNDb() error {
	geo2DBPath := filepath.Join("GeoLite2-ASN.mmdb")
	db, err := geoip2.Open(geo2DBPath)
	if err != nil {
		fmt.Errorf("Failed to open GeoLite2-ASN.mmdb: %v\n", err)
		return err
	}
	c.geoip2Db = db
	return nil
}

func (c *CDNCheck) CloseGeoASNDb() {
	if c.geoip2Db != nil {
		err := c.geoip2Db.Close()
		if err != nil {
			return
		}
	}
}
