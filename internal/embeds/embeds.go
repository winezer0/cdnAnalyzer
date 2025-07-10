package embeds

import _ "embed"

//go:embed config.yaml
var configFile string

// GetConfig 返回嵌入的配置文件内容
func GetConfig() string {
	return configFile
}
