package buffer

import (
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config Buffer Config
//
//	@author kevineluo
//	@update 2023-03-15 09:31:54
type Config struct {
	ID               string        // buffer identify ID
	ChanBufSize      int           // lock free channel size(when data in channel reach chanBufSize, it will block in Buffer.Put)
	DisableAutoFlush bool          // whether disable automate flush
	FlushInterval    time.Duration // automate flush data every [flushInterval] duration
	SyncAutoFlush    bool          // determine the buffer will automate flush asynchronously or synchronously, default is false -- async flush

	Logger   *logr.Logger // third-part logger implement logr.LogSinker, default using zapr.Logger
	LogLevel int          // used when Config.logger is nil, follow the zap style level(https://pkg.go.dev/go.uber.org/zap@v1.24.0/zapcore#Level), setting the log level for zapr.Logger(config.logLevel should be in range[-1, 5], default is 0 -- InfoLevel)
}

// Validate check config and set default value
//
//	@receiver config *Config
//	@return err error
//	@author kevineluo
//	@update 2023-03-15 10:11:28
func (config *Config) Validate() (err error) {
	if config.ID == "" {
		config.ID = uuid.NewString()
	}
	if config.ChanBufSize == 0 {
		config.ChanBufSize = 100
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 15 * time.Second
	}
	if config.Logger == nil {
		var cfg zap.Config
		level := zapcore.Level(config.LogLevel)
		if level >= zap.DebugLevel && level <= zap.FatalLevel {
			if level == zap.DebugLevel {
				cfg = zap.NewDevelopmentConfig()
			} else {
				cfg = zap.NewProductionConfig()
			}
		} else {
			err = fmt.Errorf("[config.Check] found invalid config.logLevel: %d, config.logLevel should be in range[-1, 5]", config.LogLevel)
			return
		}
		zapLogger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(cfg.EncoderConfig), zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), level))
		logger := zapr.NewLogger(zapLogger)
		config.Logger = &logger
	}
	return
}
