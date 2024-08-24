package logging

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"os"
	"sync"
)

var (
	once   sync.Once
	logger *zap.Logger
)

func IsProduction() bool {
	return os.Getenv("APP_ENV") == "production"
}

func Get() *zap.Logger {
	once.Do(func() {
		if IsProduction() {
			logger = zap.Must(zap.NewProduction())
		} else {
			logger = zap.Must(zap.NewDevelopment())
		}
	})
	return logger
}

// GetLoggerFromContext returns the logger for the given context or the default logger if not found
// :param ctx: the fiber context
// :return: the logger
func GetLoggerFromContext(ctx *fiber.Ctx) *zap.Logger {
	if l := ctx.Locals("logger"); l != nil {
		return l.(*zap.Logger)
	}
	return Get()
}
