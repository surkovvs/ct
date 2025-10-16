package ctsarama

import (
	"fmt"
	"sync"

	"github.com/IBM/sarama"
	"github.com/surkovvs/ct/internal/tools"
)

//nolint:gochecknoglobals // i'll prefer to do it once
var setLogger = &sync.Once{}

const debugTitle string = "sarama_event"

type debugLogger interface {
	Debug(msg string, args ...any)
}

func InitSaramaLogger(title string, log debugLogger) {
	setLogger.Do(func() {
		adapter := saramaLogAdapter{}
		if title != "" {
			adapter.debugTitle = title
		} else {
			adapter.debugTitle = debugTitle
		}

		if log == nil {
			log = tools.NewDefaultLogger()
		}
		adapter.log = log

		//nolint: reassign // as sarama says
		sarama.Logger = adapter
	})
}

type saramaLogAdapter struct {
	debugTitle string
	log        debugLogger
}

func (sla saramaLogAdapter) Print(v ...interface{}) {
	sla.log.Debug(
		sla.debugTitle,
		"message", fmt.Sprint(v...),
	)
}

func (sla saramaLogAdapter) Printf(format string, v ...interface{}) {
	sla.log.Debug(
		sla.debugTitle,
		"message", fmt.Sprintf(format, v...),
	)
}

func (sla saramaLogAdapter) Println(v ...interface{}) {
	sla.log.Debug(
		sla.debugTitle,
		"message", fmt.Sprint(v...),
	)
}

type loggerStub struct{}

func (loggerStub) Debug(string, ...any) {}
func (loggerStub) Info(string, ...any)  {}
func (loggerStub) Warn(string, ...any)  {}
func (loggerStub) Error(string, ...any) {}
