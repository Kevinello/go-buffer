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
	chanBufSize      int           // lock free channel size(when data in channel reach chanBufSize, it will block in Buffer.Put)
	disableAutoFlush bool          // whether disable automate flush
	flushInterval    time.Duration // automate flush data every [flushInterval] duration
	syncAutoFlush    bool          // determine the buffer will automate flush asynchronously or synchronously, default is false -- async flush

	logger   *logr.Logger // third-part logger implement logr.LogSinker, default using zapr.Logger
	logLevel int          // used when Config.logger is nil, follow the zap style level(https://pkg.go.dev/go.uber.org/zap@v1.24.0/zapcore#Level), setting the log level for zapr.Logger(config.logLevel should be in range[-1, 5], default is 0 -- InfoLevel)
}

// Check check config and set default value
//
//	@receiver config *Config
//	@return err error
//	@author kevineluo
//	@update 2023-03-15 10:11:28
func (config *Config) Check() (err error) {
	if config.ID == "" {
		config.ID = uuid.NewString()
	}
	if config.chanBufSize == 0 {
		config.chanBufSize = 100
	}
	if config.flushInterval == 0 {
		config.flushInterval = 15 * time.Second
	}
	if config.logger == nil {
		var cfg zap.Config
		level := zapcore.Level(config.logLevel)
		if level >= zap.DebugLevel && level <= zap.FatalLevel {
			if level == zap.DebugLevel {
				cfg = zap.NewDevelopmentConfig()
			} else {
				cfg = zap.NewProductionConfig()
			}
		} else {
			err = fmt.Errorf("[config.Check] found invalid config.logLevel: %d, config.logLevel should be in range[-1, 5]", config.logLevel)
			return
		}
		zapLogger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(cfg.EncoderConfig), zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), level))
		logger := zapr.NewLogger(zapLogger)
		config.logger = &logger
	}
	return
}
