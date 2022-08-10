package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var zapLog *zap.Logger

func init() {
	var err error
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	enccoderConfig := zap.NewProductionEncoderConfig()
	enccoderConfig.StacktraceKey = "" // to hide stacktrace info
	enccoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	config.EncoderConfig = enccoderConfig

	zapLog, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
}

func Info(message string, values ...interface{}) {
	zapLog.Sugar().Infof(message, values...)
}

func Debug(message string, values ...interface{}) {
	zapLog.Sugar().Debugf(message, values...)
}

func Error(message string, values ...interface{}) {
	zapLog.Sugar().Errorf(message, values...)
}

func Fatal(message string, values ...interface{}) {
	zapLog.Sugar().Fatalf(message, values...)
}

func Warning(message string, values ...interface{}) {
	zapLog.Sugar().Warnf(message, values...)
}
