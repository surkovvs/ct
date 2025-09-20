package ctsarama

import (
	"fmt"
	"sync"

	"github.com/IBM/sarama"
	"github.com/surkovvs/ct/internal/tools"
)

var (
	setLogger         = &sync.Once{}
	debugTitle string = "sarama_event"
)

type debugLogger interface {
	Debug(msg string, args ...any)
}

func InitSaramaLogger(title string, log debugLogger) {
	setLogger.Do(func() {
		if title != "" {
			debugTitle = title
		}
		if log == nil {
			log = tools.NewDefaultLogger()
		}
		sarama.Logger = saramaLogAdapter{
			log: log,
		}
	})
}

type saramaLogAdapter struct {
	log debugLogger
}

func (sla saramaLogAdapter) Print(v ...interface{}) {
	sla.log.Debug(
		debugTitle,
		"message", fmt.Sprint(v...),
	)
}

func (sla saramaLogAdapter) Printf(format string, v ...interface{}) {
	sla.log.Debug(
		debugTitle,
		"message", fmt.Sprintf(format, v...),
	)
}

func (sla saramaLogAdapter) Println(v ...interface{}) {
	sla.log.Debug(
		debugTitle,
		"message", fmt.Sprint(v...),
	)
}
