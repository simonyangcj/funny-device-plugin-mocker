package log

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitializeLogger() {
	if Logger == nil {
		Logger, _ = zap.NewProduction()
	}
}
