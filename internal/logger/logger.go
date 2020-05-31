package logger

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

const (
	_info int = iota
	_warn
	_error
	_debug
	_panic
)

var logger *zap.Logger

func init() {
	var err error
	logger, err = config.Build()

	if err != nil {
		panic(err)
	}
}

func getHeaders(ctx context.Context) *metadata.MD {
	headers, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return nil
	}

	return &headers
}

func getCorrelationID(ctx context.Context) string {
	md := getHeaders(ctx)

	if md == nil {
		return ""
	}

	values := md.Get("correlation-id")

	if len(values) > 0 {
		return values[0]
	}

	return ""
}

func log(ctx context.Context, lvl int, msg string, tags ...zap.Field) {
	correlationID := getCorrelationID(ctx)
	logFields := append(tags, zap.String("correlation-id", correlationID))

	switch lvl {
	case _info:
		logger.Info(msg, logFields...)
	case _warn:
		logger.Warn(msg, logFields...)
	case _error:
		logger.Error(msg, logFields...)
	case _debug:
		logger.Error(msg, logFields...)
	case _panic:
		logger.Panic(msg, logFields...)
	}

	_ = logger.Sync() // https://github.com/uber-go/zap/issues/370
}

// Info logs an info statement from the application.
func Info(ctx context.Context, msg string, tags ...zap.Field) {
	log(ctx, _info, msg, tags...)
}

// Warn logs an info statement from the application.
func Warn(ctx context.Context, msg string, tags ...zap.Field) {
	log(ctx, _warn, msg, tags...)
}

// Error logs an info statement from the application.
func Error(ctx context.Context, msg string, tags ...zap.Field) {
	log(ctx, _error, msg, tags...)
}

// Debug logs an info statement from the application.
func Debug(ctx context.Context, msg string, tags ...zap.Field) {
	log(ctx, _debug, msg, tags...)
}

// Panic logs an info statement from the application.
func Panic(ctx context.Context, msg string, tags ...zap.Field) {
	log(ctx, _panic, msg, tags...)
}
