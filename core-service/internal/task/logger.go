package task

import (
	"fmt"

	"github.com/Djarottosca/finalproject-ftgo-1/pkg/logger"
)

// asynqLogger adapts pkg/logger's zerolog global logger to asynq.Logger, so
// the worker's internal logs (job start/retry/failure) match the rest of
// core-service's log format instead of asynq's default stdlib logger.
type asynqLogger struct{}

// NewAsynqLogger returns an asynq.Logger backed by pkg/logger.
func NewAsynqLogger() *asynqLogger { return &asynqLogger{} }

func (l *asynqLogger) Debug(args ...interface{}) { logger.Log.Debug().Msg(fmt.Sprint(args...)) }
func (l *asynqLogger) Info(args ...interface{})  { logger.Log.Info().Msg(fmt.Sprint(args...)) }
func (l *asynqLogger) Warn(args ...interface{})  { logger.Log.Warn().Msg(fmt.Sprint(args...)) }
func (l *asynqLogger) Error(args ...interface{}) { logger.Log.Error().Msg(fmt.Sprint(args...)) }
func (l *asynqLogger) Fatal(args ...interface{}) { logger.Log.Fatal().Msg(fmt.Sprint(args...)) }
