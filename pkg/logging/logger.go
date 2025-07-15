package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	sugar  *zap.SugaredLogger
)

// LogConfig 日志配置选项
type LogConfig struct {
	LogLevel     string // 日志级别
	LogFile      string // 日志输出文件
	StdoutPrefix bool   // 是否在控制台输出显示日志前缀（时间戳、级别等）
}

// InitLogger 初始化日志系统，支持 stdout 和文件同时输出
// 默认控制台输出不显示前缀，文件输出保留完整信息
func InitLogger(logLevel string, logFile string) error {
	config := LogConfig{
		LogLevel:     logLevel,
		LogFile:      logFile,
		StdoutPrefix: false, // 默认控制台不显示前缀
	}
	return InitLoggerWithConfig(config)
}

// InitLoggerWithConfig 使用配置初始化日志系统
func InitLoggerWithConfig(config LogConfig) error {
	var level zapcore.Level
	switch strings.ToLower(config.LogLevel) {
	case "debug":
		level = zapcore.DebugLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 控制台编码器配置
	consoleEncoderConfig := zap.NewProductionEncoderConfig()
	consoleEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 控制台输出不显示前缀
	if !config.StdoutPrefix {
		consoleEncoderConfig.TimeKey = ""
		consoleEncoderConfig.LevelKey = ""
		consoleEncoderConfig.CallerKey = ""
		consoleEncoderConfig.FunctionKey = ""
	}

	// 文件编码器配置 - 始终保留完整信息
	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var cores []zapcore.Core

	// 控制台输出
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderConfig),
		zapcore.Lock(os.Stdout),
		level,
	)
	cores = append(cores, consoleCore)

	// 文件输出
	if config.LogFile != "" {
		if err := ensureLogDir(config.LogFile); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		file, err := os.OpenFile(config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}

		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(fileEncoderConfig),
			zapcore.AddSync(file),
			level,
		)
		cores = append(cores, fileCore)
	}

	core := zapcore.NewTee(cores...)

	// 添加调用者信息，并设置跳过级别
	opts := []zap.Option{zap.AddCaller(), zap.AddCallerSkip(1)}

	logger = zap.New(core, opts...)
	sugar = logger.Sugar()

	sugar.Infow("Logger initialized",
		"level", config.LogLevel,
		"output", config.LogFile,
	)

	return nil
}

// Sync 刷新日志缓冲区到磁盘
func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}

// 辅助函数（保留你想要的封装）
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

// 辅助函数：确保日志目录存在
func ensureLogDir(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, 0755)
}
