package middlewares

import (
	"fmt"
	"github.com/darmiel/ralf/cmd/server/logging"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"time"
)

const (
	CorrelationIDHeader = "X-Correlation-ID"
	CorrelationIDKey    = "X-Correlation-ID"
)

// LogMiddleware is a middleware that logs the request and response
// :param ctx: the fiber context
// :return: the next handler
func LogMiddleware(ctx *fiber.Ctx) error {
	// create a correlation ID to trace requests
	correlationID := xid.New().String()
	ctx.Set(CorrelationIDHeader, correlationID)
	ctx.Locals(CorrelationIDKey, correlationID) // for convenience

	// create a log
	//ger with the correlation ID and store it in the context
	// so that it can be used in other middlewares and handlers by calling GetLoggerFromContext
	// or by using ctx.Locals("logger").(*zap.Logger)
	logger := logging.Get().With(zap.String(CorrelationIDKey, correlationID))
	ctx.Locals("logger", logger)

	// track time the request takes to complete
	startTime := time.Now()

	res := ctx.Next()

	// get status code from response (if available)
	response := ctx.Response()
	statusCode := -1
	if response != nil {
		statusCode = response.StatusCode()
	}

	// get logger from locals again in case it was updated by another middleware
	logger = ctx.Locals("logger").(*zap.Logger)

	logger.Info(
		fmt.Sprintf(
			"%s request to %s completed",
			ctx.Method(),
			ctx.OriginalURL(),
		),
		zap.String("method", ctx.Method()),
		zap.String("url", ctx.OriginalURL()),
		zap.String("user_agent", ctx.Get("User-Agent")),
		zap.Int("status_code", statusCode),
		zap.Duration("elapsed_ms", time.Since(startTime)),
	)

	return res
}
