package logging

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func InitializeLogging(configuration *LogConfiguration, option ...zap.Option) error {
	cores := make([]zapcore.Core, len(configuration.Builders))
	var index int
	for _, builder := range configuration.Builders {
		core, err := builder.Build()
		if err != nil {
			return err
		}
		cores[index] = core
		index++
	}
	core := zapcore.NewTee(cores...)

	logger := zap.New(core, zap.AddCaller())

	if option != nil {
		logger = logger.WithOptions(option...)
	}
	log = logger
	return nil
}

type LoggerID string

const (
	loggerID LoggerID = "loggerID"
)

func GetLogger() *zap.Logger {
	if log == nil {
		panic("Logger not initialized")
	}
	return log
}
func Logger(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(string(loggerID)).(*zap.Logger)
	if !ok {
		logger = GetLogger()
	}
	return logger
}
func SetLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerID, l)
}

type LogConfiguration struct {
	Builders map[string]CoreBuilder
}

type CoreBuilder interface {
	Build() (zapcore.Core, error)
}

type ConsoleBuilder struct {
	encoder zapcore.Encoder
	level   zapcore.Level
}

func NewConsoleBuilder(encConfig zapcore.EncoderConfig, level zapcore.Level) *ConsoleBuilder {
	return &ConsoleBuilder{encoder: zapcore.NewConsoleEncoder(encConfig), level: level}
}

func (cb *ConsoleBuilder) Build() (zapcore.Core, error) {
	return zapcore.NewCore(cb.encoder, zapcore.AddSync(os.Stdout), cb.level), nil
}

type FileBuilder struct {
	encoder zapcore.Encoder
	level   zapcore.Level
	path    string
}

func NewFileBuilder(path string, encConfig zapcore.EncoderConfig, level zapcore.Level) *FileBuilder {
	return &FileBuilder{encoder: zapcore.NewJSONEncoder(encConfig), level: level, path: path}
}

func (cb *FileBuilder) Build() (zapcore.Core, error) {
	file, err := os.OpenFile(cb.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return zapcore.NewCore(cb.encoder, zapcore.AddSync(file), cb.level), nil
}
