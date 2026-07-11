package middleware

import (
	echo "github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/Djarottosca/finalproject-ftgo-1/pkg/logger"
)

func RequestLoggerMiddleware() echo.MiddlewareFunc {
	return echoMiddleware.RequestLoggerWithConfig(echoMiddleware.RequestLoggerConfig{
		LogLatency:   true,
		LogRemoteIP:  true,
		LogMethod:    true,
		LogURI:       true,
		LogRequestID: true,
		LogUserAgent: true,
		LogStatus:    true,
		LogError:     true,
		LogValuesFunc: func(c echo.Context, v echoMiddleware.RequestLoggerValues) error {
			if v.Error != nil {
				logger.Log.Error().
					Str("request_id", v.RequestID).
					Str("method", v.Method).
					Str("uri", v.URI).
					Int("status", v.Status).
					Str("ip", v.RemoteIP).
					Str("user_agent", v.UserAgent).
					Str("latency", v.Latency.String()).
					Err(v.Error).
					Msg("request")
			} else {
				logger.Log.Info().
					Str("request_id", v.RequestID).
					Str("method", v.Method).
					Str("uri", v.URI).
					Int("status", v.Status).
					Str("ip", v.RemoteIP).
					Str("user_agent", v.UserAgent).
					Str("latency", v.Latency.String()).
					Msg("request")
			}
			return nil
		},
	})
}
