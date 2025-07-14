package log

import (
	"os"

	"github.com/maahdima/mwp/api/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(cfg config.AppConfig) {
	stdout := zapcore.AddSync(os.Stdout)

	logLevel := zap.WarnLevel
	if cfg.Mode == "development" {
		logLevel = zap.DebugLevel
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	var consoleEncoder zapcore.Encoder
	if cfg.ConsoleLogFormat == "json" {
		consoleEncoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		consoleEncoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, zap.NewAtomicLevelAt(logLevel)),
	)

	zap.ReplaceGlobals(zap.New(core, zap.AddCaller()))
}
