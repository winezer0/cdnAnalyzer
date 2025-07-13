package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

func InitLogger(logLevel, logOutput string) error {
	var level zapcore.Level
	if err := level.Set(logLevel); err != nil {
		return err
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var core zapcore.Core
	if logOutput == "stdout" || logOutput == "" {
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)
	} else {
		file, err := os.OpenFile(logOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(file),
			level,
		)
	}

	Logger = zap.New(core, zap.AddCaller())
	Sugar = Logger.Sugar()
	return nil
}

func Debugf(template string, args ...interface{}) {
	Sugar.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	Sugar.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	Sugar.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	Sugar.Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	Sugar.Fatalf(template, args...)
}

func Debugln(args ...interface{}) {
	Sugar.Debugln(args...)
}

func Debug(args ...interface{}) {
	Sugar.Debug(args...)
}

func Info(args ...interface{}) {
	Sugar.Info(args...)
}

func Warn(args ...interface{}) {
	Sugar.Warn(args...)
}

func Error(args ...interface{}) {
	Sugar.Error(args...)
}

func Fatal(args ...interface{}) {
	Sugar.Fatal(args...)
}
