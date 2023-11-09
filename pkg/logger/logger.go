package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/kcloutie/knot/pkg/params/settings"
	"github.com/kcloutie/knot/pkg/params/version"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ctxKey struct{}

var (
	once   sync.Once
	logger *zap.Logger
)

const (
	RootCommandKey = "root_command"
	SubCommandKey  = "sub_command"
	CommitKey      = "commit"
	VersionKey     = "version"
	BuildTimeKey   = "build_time"
	GoVersionKey   = "go_version"
	TimeStampKey   = "timestamp"
	MessageKey     = "message"
	DurationKey    = "duration"
	UrlKey         = "url"
	EnvKey         = "environment"
)

// logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

// Get initializes a zap.Logger instance if it has not been initialized
// already and returns the same instance for subsequent calls.
func Get() *zap.Logger {
	once.Do(func() {
		stdout := zapcore.AddSync(os.Stdout)

		file := zapcore.AddSync(&lumberjack.Logger{
			Filename:   fmt.Sprintf("%s.log", settings.CliBinaryName),
			MaxSize:    5,
			MaxBackups: 10,
			MaxAge:     14,
			Compress:   true,
		})

		level := zap.InfoLevel
		levelEnv := os.Getenv("LOG_LEVEL")
		if levelEnv != "" {
			levelFromEnv, err := zapcore.ParseLevel(levelEnv)
			if err != nil {
				log.Println(
					fmt.Errorf("invalid level, defaulting to INFO: %w", err),
				)
			}

			level = levelFromEnv
		} else {
			if settings.DebugModeEnabled {
				level = zap.DebugLevel
			}
		}

		logLevel := zap.NewAtomicLevelAt(level)

		productionCfg := zap.NewProductionEncoderConfig()
		productionCfg.TimeKey = TimeStampKey
		productionCfg.MessageKey = MessageKey
		productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

		developmentCfg := zap.NewDevelopmentEncoderConfig()
		developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

		cldLog := os.Getenv("CLD_LOG")
		logToConsole := strings.Contains(strings.ToUpper(cldLog), "CONSOLE")
		if settings.DebugModeEnabled {
			logToConsole = true
		}
		logToFile := strings.Contains(strings.ToUpper(cldLog), "FILE")

		consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
		fileEncoder := zapcore.NewJSONEncoder(productionCfg)

		buildInfo, _ := debug.ReadBuildInfo()

		var core zapcore.Core
		if !logToConsole && !logToFile {
			core = zapcore.NewTee()
		}
		if logToConsole && logToFile {
			core = zapcore.NewTee(
				zapcore.NewCore(consoleEncoder, stdout, logLevel),
				zapcore.NewCore(fileEncoder, file, logLevel).
					With(
						[]zapcore.Field{
							zap.String(CommitKey, version.Commit),
							zap.String(VersionKey, version.BuildVersion),
							zap.String(BuildTimeKey, version.BuildTime),
							zap.String(GoVersionKey, buildInfo.GoVersion),
						},
					),
			)
		} else {
			if logToFile {
				core = zapcore.NewTee(
					zapcore.NewCore(fileEncoder, file, logLevel).
						With(
							[]zapcore.Field{
								zap.String(CommitKey, version.Commit),
								zap.String(VersionKey, version.BuildVersion),
								zap.String(BuildTimeKey, version.BuildTime),
								zap.String(GoVersionKey, buildInfo.GoVersion),
							},
						),
				)
			}
			if logToConsole {
				core = zapcore.NewTee(
					zapcore.NewCore(consoleEncoder, stdout, logLevel),
				)
			}
		}

		logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.FatalLevel))
	})

	return logger
}

// FromCtx returns the Logger associated with the ctx. If no logger
// is associated, the default logger is returned, unless it is nil
// in which case a disabled logger is returned.
func FromCtx(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		return l
	} else if l := logger; l != nil {
		return l
	}
	return zap.NewNop()
}

// WithCtx returns a copy of ctx with the Logger attached.
func WithCtx(ctx context.Context, l *zap.Logger) context.Context {
	if lp, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		if lp == l {
			return ctx
		}
	}
	return context.WithValue(ctx, ctxKey{}, l)
}

type LeveledLogger struct {
	logger *zap.Logger
}

func NewLeveledLogger(lgr *zap.Logger) LeveledLogger {
	l := LeveledLogger{logger: lgr.WithOptions(zap.AddCallerSkip(3))}
	return l
}

func (l *LeveledLogger) Error(msg string, keysAndValues ...interface{}) {
	lgr := addFields(l.logger, keysAndValues)
	lgr.Level()
	lgr.Error(msg)
}

func (l *LeveledLogger) Info(msg string, keysAndValues ...interface{}) {
	lgr := addFields(l.logger, keysAndValues)
	lgr.Info(msg)
}

func (l *LeveledLogger) Debug(msg string, keysAndValues ...interface{}) {
	lgr := addFields(l.logger, keysAndValues)
	lgr.Debug(msg)
}

func (l *LeveledLogger) Warn(msg string, keysAndValues ...interface{}) {
	lgr := addFields(l.logger, keysAndValues)
	lgr.Warn(msg)
}

func addFields(lgr *zap.Logger, keysAndValues []interface{}) *zap.Logger {
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		lgr = lgr.With(zap.String(keysAndValues[i].(string), fmt.Sprintf("%v", keysAndValues[i+1])))
	}
	return lgr
}
