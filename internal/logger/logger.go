package logger

import (
	"github.com/Filiphasan/rabbitmqgolang/configs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func UseLogger(appConfig *configs.AppConfig) *zap.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "@timestamp"
	encoderConfig.LevelKey = "LogLevel"
	encoderConfig.MessageKey = "Message"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.AddSync(os.Stdout), zap.InfoLevel)

	logger := zap.New(core)
	logger = logger.With(zap.String("ProjectName", appConfig.ProjectName), zap.String("Environment", appConfig.Environment))

	zap.ReplaceGlobals(logger)
	zap.RedirectStdLog(logger)

	return logger
}
