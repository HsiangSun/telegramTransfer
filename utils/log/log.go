package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"telgramTransfer/utils/config"
	"time"
)

var Sugar *zap.SugaredLogger

// InitLog 日志初始化
func InitLog() {
	writeSyncer := getLogWriter()
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	logger := zap.New(core, zap.AddCaller())
	Sugar = logger.Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	logPath := fmt.Sprintf("%s%s%s%s", config.AppPath, "log", string(os.PathSeparator), "log")
	file := fmt.Sprintf("%s/log_%s.log",
		logPath,
		time.Now().Format("20060102"))
	lumberJackLogger := &lumberjack.Logger{
		Filename:   file,
		MaxSize:    config.LogC.MaxSize,
		MaxBackups: config.LogC.MaxBackups,
		MaxAge:     config.LogC.MaxAge,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}
