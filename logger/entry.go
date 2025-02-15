package logger

import (
	"fmt"
	"time"
)

type Entry struct {
	logger  *Logger
	limit   time.Duration
	level   Level
	message string
	time    time.Time
	data    Fields
}

func (e *Entry) Panic(args ...interface{}) {
	e.log(PanicLevel, args...)
}

func (e *Entry) Crit(args ...interface{}) {
	e.log(CritLevel, args...)
}

func (e *Entry) Error(args ...interface{}) {
	e.log(ErrorLevel, args...)
}

func (e *Entry) Warn(args ...interface{}) {
	e.log(WarnLevel, args...)
}

func (e *Entry) Info(args ...interface{}) {
	e.log(InfoLevel, args...)
}

func (e *Entry) Debug(args ...interface{}) {
	e.log(DebugLevel, args...)
}

func (e *Entry) Trace(args ...interface{}) {
	e.log(TraceLevel, args...)
}

func (e *Entry) Limit(dur time.Duration) *Entry {
	e.limit = dur
	return e
}

func (e *Entry) log(level Level, args ...interface{}) {
	if e.limit != 0 {
		token := ""
		if len(args) > 0 {
			if str, ok := args[0].(string); ok {
				token = str
			}
		}

		e.logger.limitsLock.Lock()
		timestamp := e.logger.limits[token]
		if time.Since(timestamp) < e.limit {
			e.logger.limitsLock.Unlock()
			return
		}

		e.logger.limits[token] = time.Now()
		e.logger.limitsLock.Unlock()
	}

	if time.Since(e.logger.limitsClean) > e.logger.maxLimit {
		e.logger.cleanLimits()
	}

	e.level = level
	e.message = fmt.Sprint(args...)
	e.time = time.Now()

	rec := &Record{
		Level:   e.level,
		Message: e.message,
		Time:    e.time,
		Data:    e.data,
		logger:  e.logger,
	}

	for _, hand := range e.logger.handlers {
		hand(rec)
	}
}
