package downfile

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

// ClientConfig HTTP客户端配置
type ClientConfig struct {
	ConnectTimeout int    // 连接超时时间（秒）
	IdleTimeout    int    // 空闲超时时间（秒）
	ProxyURL       string // 代理URL（支持http和socks5）
}

// DefaultClientConfig 返回默认的HTTP客户端配置
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		ConnectTimeout: 30, // 默认30秒连接超时
		IdleTimeout:    60, // 默认60秒空闲超时
		ProxyURL:       "", // 默认不使用代理
	}
}

// CreateHTTPClient 创建HTTP客户端
func CreateHTTPClient(config *ClientConfig) (*http.Client, error) {
	// 使用默认配置（如果传入nil）
	if config == nil {
		config = DefaultClientConfig()
	}

	// 设置代理
	var proxyFunc func(*http.Request) (*url.URL, error)
	if config.ProxyURL != "" {
		proxyURL, err := url.Parse(config.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("解析代理URL失败: %w", err)
		}
		proxyFunc = http.ProxyURL(proxyURL)
	} else {
		proxyFunc = http.ProxyFromEnvironment
	}

	// 配置传输设置
	transport := &http.Transport{
		Proxy: proxyFunc,
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(config.ConnectTimeout) * time.Second, // 连接超时
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       time.Duration(config.IdleTimeout) * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: time.Duration(config.ConnectTimeout) * time.Second,
	}

	// 创建HTTP客户端
	httpClient := &http.Client{
		Transport: transport,
		// 不设置整体超时，使用context控制
	}

	return httpClient, nil
}
