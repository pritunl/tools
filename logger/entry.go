package logger

import (
	"fmt"
	"time"
)

// Entry is a pending log message being assembled before it is emitted. An
// Entry is created by Logger.WithFields or the package-level WithFields,
// optionally rate-limited with Limit, and then emitted by calling one of the
// level methods (Info, Error, etc.). An Entry is not safe for concurrent use
// and should not be reused after it has been emitted.
type Entry struct {
	logger  *Logger
	limit   time.Duration
	level   Level
	message string
	time    time.Time
	data    Fields
}

// Panic emits the entry at PanicLevel. The message is formed from args in
// the manner of fmt.Sprint. It does not call the built-in panic.
func (e *Entry) Panic(args ...interface{}) {
	e.log(PanicLevel, args...)
}

// Crit emits the entry at CritLevel. The message is formed from args in the
// manner of fmt.Sprint.
func (e *Entry) Crit(args ...interface{}) {
	e.log(CritLevel, args...)
}

// Error emits the entry at ErrorLevel. The message is formed from args in
// the manner of fmt.Sprint.
func (e *Entry) Error(args ...interface{}) {
	e.log(ErrorLevel, args...)
}

// Warn emits the entry at WarnLevel. The message is formed from args in the
// manner of fmt.Sprint.
func (e *Entry) Warn(args ...interface{}) {
	e.log(WarnLevel, args...)
}

// Info emits the entry at InfoLevel. The message is formed from args in the
// manner of fmt.Sprint.
func (e *Entry) Info(args ...interface{}) {
	e.log(InfoLevel, args...)
}

// Debug emits the entry at DebugLevel. The message is formed from args in
// the manner of fmt.Sprint.
func (e *Entry) Debug(args ...interface{}) {
	e.log(DebugLevel, args...)
}

// Trace emits the entry at TraceLevel. The message is formed from args in
// the manner of fmt.Sprint.
func (e *Entry) Trace(args ...interface{}) {
	e.log(TraceLevel, args...)
}

// Limit rate-limits the entry so that a message with the same text is
// emitted at most once per dur. Rate limiting is keyed on the first argument
// passed to the level method when it is a string; all other arguments and
// the fields are ignored for the purpose of matching. Subsequent emissions of
// the same message within dur are silently dropped.
//
// dur should not exceed the logger's maximum limit (see SetMaxLimit), which
// bounds how long the logger remembers when a message was last emitted.
// Limit returns e to allow chaining.
func (e *Entry) Limit(dur time.Duration) *Entry {
	e.limit = dur
	return e
}

// log applies rate limiting, triggers periodic cleanup of expired rate-limit
// entries, and dispatches a Record built from the entry to every handler
// registered on the logger.
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
