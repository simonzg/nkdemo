package simulator

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"

	"fmt"
)

type loggerEnabler struct {
	verbose bool
}

func (l *loggerEnabler) Enabled(level zapcore.Level) bool {
	return l.verbose || level > zapcore.DebugLevel
}

func NewJSONLogger(output *os.File, verbose bool) *zap.Logger {
	jsonEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	core := zapcore.NewCore(jsonEncoder, output, &loggerEnabler{verbose})
	options := []zap.Option{zap.AddStacktrace(zap.ErrorLevel)}

	return zap.New(core, options...)
}

func NewLogger(logDir, id string) *zap.Logger {
	output := os.Stdout
	err := os.MkdirAll(filepath.FromSlash(logDir), 0755)
	if err != nil {
		fmt.Println("Could not create log directory", err)
		return nil
	}

	output, err = os.Create(filepath.FromSlash(fmt.Sprintf("%s/%s.log", logDir, id)))
	if err != nil {
		fmt.Println("Could not create log file", zap.Error(err))
		return nil
	}

	logger := NewJSONLogger(output, true)
	logger = logger.With(zap.String("id", id))

	return logger
}
