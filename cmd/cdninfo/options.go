package main

// 版本信息常量（根据实际情况修改）
const (
	AppName          = "cdninfo"
	AppShortDesc     = "CDN Information Analysis Tool"
	AppLongDesc      = "CDN Information Analysis Tool, Analysis Such as (Domain resolution|IP analysis|CDN|WAF|Cloud)."
	DefaultConfigUrl = "https://raw.githubusercontent.com/winezer0/cdninfo/refs/heads/main/cmd/cdninfo/cdninfo.yaml"
	AppVersion       = "0.6.1"
	BuildDate        = "2026-04-26"
)

// CmdConfig 存储程序配置，使用结构体标签定义命令行参数
type CmdConfig struct {
	// 配置文件参数
	ConfigFile   string `short:"c" long:"config-file" description:"config yaml file path (default ~/isecdb/cdninfo.yaml)" default:""`
	UpdateConfig bool   `short:"C" long:"update-config" description:"updated default config content by remote url (default false)"`

	// 基本参数 覆盖app Config中的配置
	Input     string `short:"i" long:"input" description:"input file or str list (separated by commas)"`
	InputType string `short:"I" long:"input-type" description:"input data type: str/file/sys (default str)" default:"str" choice:"file" choice:"str" choice:"sys"`

	// 输出配置参数 覆盖app Config中的配置
	Output      string `short:"o" long:"output" description:"output file path (default result.json)" default:"result.json"`
	OutputType  string `short:"O" long:"output-type" description:"output file type: csv/json/txt/sys (default sys)" default:"sys" choice:"csv" choice:"json" choice:"txt" choice:"sys"`
	OutputLevel int    `short:"l" long:"output-level" description:"Output verbosity level: 1=quiet, 2=default, 3=detail (default 2)" default:"2" choice:"1" choice:"2" choice:"3"`
	OutputNoCDN bool   `short:"n" long:"output-no-cdn" description:"only output Info where not CDN and not WAF."`

	// 数据库更新配置
	Proxy    string `short:"p" long:"proxy" description:"use the proxy URL down files (support http|socks5)" default:""`
	Folder   string `short:"d" long:"folder" description:"db files storage dir (default ~/isecdb)" default:""`
	UpdateDB bool   `short:"u" long:"update-db" description:"Auto update db files by interval (default: false)"`

	// DNS 相关参数（新增）
	QueryMethod     string `short:"q" long:"query-method" description:"Cover Config, Set dns query method:(allow:|dns|edns|both)" default:"" choice:"" choice:"dns" choice:"edns" choice:"both"`
	DNSTimeout      int    `short:"t" long:"dns-timeout" description:"Cover Config, Set DNS query timeout in seconds" default:"0"`
	ResolversNum    int    `short:"r" long:"resolvers-num" description:"Cover Config, Set number of resolvers to use" default:"0"`
	CityMapNum      int    `short:"m" long:"city-map-num" description:"Cover Config, Set number of city map workers" default:"0"`
	DNSConcurrency  int    `short:"w" long:"dns-concurrency" description:"Cover Config, Set concurrent DNS queries" default:"0"`
	EDNSConcurrency int    `short:"W" long:"edns-concurrency" description:"Cover Config, Set concurrent EDNS queries" default:"0"`

	// 版本号输出
	Version bool `short:"v" long:"version" description:"Show Program version and exit (default: false)"`

	// 日志配置参数
	LogFile       string `long:"lf" description:"log file path (default: only stdout)" default:""`
	LogLevel      string `long:"ll" description:"log level: debug/info/warn/error (default error)" default:"error" choice:"debug" choice:"info" choice:"warn" choice:"error"`
	ConsoleFormat string `long:"lc" description:"log console format, multiple choice T(time),L(level),C(caller),F(func),M(msg). Empty or off will disable." default:"T L C M"`
}
