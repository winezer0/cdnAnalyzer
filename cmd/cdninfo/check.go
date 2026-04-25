package main

import (
	"fmt"
	"github.com/winezer0/cdninfo/internal/config"
	"github.com/winezer0/cdninfo/pkg/fileutils"
	"strings"
)

// 使用命令行非默认值参数更新配置文件中的参数
func updateAppConfigByOpts(appConfig *config.AppConfig, cmdConfig *Options) *config.AppConfig {
	if cmdConfig.DNSTimeout > 0 {
		appConfig.DNSTimeOut = cmdConfig.DNSTimeout
	}

	if cmdConfig.ResolversNum > 0 {
		appConfig.ResolversNum = cmdConfig.ResolversNum
	}

	if cmdConfig.CityMapNum > 0 {
		appConfig.CityMapNUm = cmdConfig.CityMapNum
	}

	if cmdConfig.DNSConcurrency > 0 {
		appConfig.DNSConcurrency = cmdConfig.DNSConcurrency
	}

	if cmdConfig.EDNSConcurrency > 0 {
		appConfig.EDNSConcurrency = cmdConfig.EDNSConcurrency
	}

	if cmdConfig.QueryMethod != "" {
		appConfig.QueryMethod = cmdConfig.QueryMethod
	}

	// 确保并发数有一个合理的默认值，防止死锁
	if appConfig.DNSConcurrency <= 0 {
		appConfig.DNSConcurrency = 10
	}
	if appConfig.EDNSConcurrency <= 0 {
		appConfig.EDNSConcurrency = 10
	}
	return appConfig
}

func parseTargetFomat(target string, targetType string) ([]string, error) {
	var (
		targets []string
		err     error
	)

	if target == "" && (targetType == "str" || targetType == "file") {
		return nil, fmt.Errorf("the target must be specified")
	}

	switch targetType {
	case "str":
		targets = strings.Split(target, ",")
	case "file":
		targets, err = fileutils.ReadTextToList(target)
		if err != nil {
			return nil, fmt.Errorf("failed to load the target file: %v", err)
		}
	case "sys":
		targets, err = fileutils.ReadPipeToList()
		if err != nil {
			return nil, fmt.Errorf("failed to load the system pipe: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported target type: %s", targetType)
	}

	// 过滤空字符串
	filtered := make([]string, 0, len(targets))
	for _, t := range targets {
		t = strings.TrimSpace(t)
		if t != "" {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no valid target has been entered")
	}

	return filtered, nil
}
