package buffer

import (
	"time"

	"github.com/go-logr/logr"
)

// Config Buffer Config
//
//	@author kevineluo
//	@update 2023-03-15 09:31:54
type Config struct {
	chanBufSize   int           // lock free channel size(when data in channel reach chanBufSize, it will block in Buffer.Put)
	flushInterval time.Duration // automate flush data every [flushInterval] duration

	logger *logr.Logger // logger implement logr.LogSinker
}

// Check check config and set default value
//
//	@receiver config *Config
//	@return err error
//	@author kevineluo
//	@update 2023-03-15 10:11:28
func (config *Config) Check() (err error) {
	if config.chanBufSize == 0 {
		config.chanBufSize = 100
	}
	if config.flushInterval == 0 {
		config.flushInterval = 15 * time.Second
	}
	if config.logger == nil {
		logger := logr.Discard()
		config.logger = &logger
	}
	return
}
