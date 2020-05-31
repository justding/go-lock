// +build release

package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var config = zap.Config{
	OutputPaths: []string{"stdout"},
	Encoding:    "json",
	Level:       zap.NewAtomicLevel(),
	EncoderConfig: zapcore.EncoderConfig{
		MessageKey:   "msg",
		LevelKey:     "level",
		TimeKey:      "time",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeLevel:  zapcore.LowercaseLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	},
}
