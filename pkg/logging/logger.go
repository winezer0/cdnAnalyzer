package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 全局变量
var (
	logger *zap.Logger
	sugar  *zap.SugaredLogger
)

// LogConfig 日志配置结构体（所有字段都不导出）
type LogConfig struct {
	logLevel      string // 日志级别 debug/info/warn/error
	logFile       string // 日志文件路径（为空则不写入文件）
	consoleFormat string // 控制台格式字符串，空或 "off" 表示关闭控制台输出
}

// NewLogConfig 创建一个新的日志配置
func NewLogConfig(logLevel, logFile, consoleFormat string) LogConfig {
	return LogConfig{
		logLevel:      logLevel,
		logFile:       logFile,
		consoleFormat: consoleFormat,
	}
}

func (c LogConfig) LogLevel() string {
	return c.logLevel
}

func (c LogConfig) LogFile() string {
	return c.logFile
}

func (c LogConfig) ConsoleFormat() string {
	return c.consoleFormat
}

// parseEncoderConfig 解析格式字符串生成对应的 EncoderConfig
func parseEncoderConfig(format string) zapcore.EncoderConfig {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncodeLevel = zapcore.CapitalLevelEncoder

	cfg.TimeKey = "time"
	cfg.LevelKey = "level"
	cfg.CallerKey = "caller"
	cfg.MessageKey = "msg"
	cfg.FunctionKey = "function"

	showTime := strings.Contains(format, "T")
	showLevel := strings.Contains(format, "L")
	showCaller := strings.Contains(format, "C")
	showMessage := strings.Contains(format, "M")
	showFunction := strings.Contains(format, "F")

	if !showTime {
		cfg.TimeKey = ""
	}
	if !showLevel {
		cfg.LevelKey = ""
	}
	if !showCaller {
		cfg.CallerKey = ""
	}
	if !showFunction {
		cfg.FunctionKey = ""
	}
	if !showMessage {
		cfg.MessageKey = ""
	}

	return cfg
}

// InitLogger 初始化日志系统
func InitLogger(config LogConfig) error {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(config.LogLevel())); err != nil {
		level = zapcore.InfoLevel
	}

	var cores []zapcore.Core

	// 控制台输出（consoleFormat 非空且非 "off" 时启用）
	if format := config.ConsoleFormat(); format != "" && format != "off" {
		encoderCfg := parseEncoderConfig(format)
		consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)

		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.Lock(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// 文件输出
	if file := config.LogFile(); file != "" {
		if err := ensureLogDir(file); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}

		fileEncoderCfg := zap.NewProductionEncoderConfig()
		fileEncoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		fileEncoder := zapcore.NewJSONEncoder(fileEncoderCfg)

		fileCore := zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(f),
			level,
		)
		cores = append(cores, fileCore)
	}

	if len(cores) == 0 {
		return fmt.Errorf("no log output target configured")
	}

	core := zapcore.NewTee(cores...)

	// 添加调用者信息
	opts := []zap.Option{zap.AddCaller(), zap.AddCallerSkip(1)}

	logger = zap.New(core, opts...)
	sugar = logger.Sugar()

	sugar.Infow("Logger initialized",
		"level", config.LogLevel(),
		"console_format", config.ConsoleFormat(),
		"file_output", config.LogFile(),
	)

	return nil
}

// Sync 刷新日志缓冲区到磁盘
func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}

// Sugar 获取 SugaredLogger
func Sugar() *zap.SugaredLogger {
	return sugar
}

// ensureLogDir 确保日志文件所在目录存在
func ensureLogDir(filePath string) error {
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func Debugf(template string, args ...interface{}) {
	if sugar != nil {
		sugar.Debugf(template, args...)
	}
}

func Infof(template string, args ...interface{}) {
	if sugar != nil {
		sugar.Infof(template, args...)
	}
}

func Warnf(template string, args ...interface{}) {
	if sugar != nil {
		sugar.Warnf(template, args...)
	}
}

func Errorf(template string, args ...interface{}) {
	if sugar != nil {
		sugar.Errorf(template, args...)
	}
}

func Fatalf(template string, args ...interface{}) {
	if sugar != nil {
		sugar.Fatalf(template, args...)
	}
}

func Debug(args ...interface{}) {
	if sugar != nil {
		sugar.Debug(args...)
	}
}

func Info(args ...interface{}) {
	if sugar != nil {
		sugar.Info(args...)
	}
}

func Warn(args ...interface{}) {
	if sugar != nil {
		sugar.Warn(args...)
	}
}

func Error(args ...interface{}) {
	if sugar != nil {
		sugar.Error(args...)
	}
}

func Fatal(args ...interface{}) {
	if sugar != nil {
		sugar.Fatal(args...)
	}
}
