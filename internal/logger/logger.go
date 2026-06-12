package logger

import (
	"strings"

	"go.uber.org/zap"
)

func New(environment string) (*zap.Logger, error) {
	if strings.EqualFold(environment, "production") {
		return zap.NewProduction()
	}

	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	return config.Build()
}
