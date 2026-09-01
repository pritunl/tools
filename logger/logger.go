// Package logger implements a leveled, structured logger with optional
// rate limiting and pluggable output handlers.
//
// The logger itself does not write anywhere. Every emitted message is turned
// into a Record and passed to each handler registered with AddHandler; a
// handler typically formats the record with Record.String or
// Record.StringColor and writes it to a file or terminal.
//
// A package-level global logger is available through the package functions
// (Info, Error, WithFields, ...) and is configured with Init. Independent
// loggers can be created with New.
//
//	logger.Init(logger.SetIcons(false))
//	logger.AddHandler(func(rec *logger.Record) {
//		fmt.Print(rec.StringColor())
//	})
//
//	logger.WithFields(logger.Fields{
//		"user": "alice",
//		"error": err,
//	}).Error("auth: Login failed")
//
// Messages can be rate-limited so that a repeated message is emitted at most
// once per interval:
//
//	logger.WithFields(fields).Limit(30 * time.Second).Warn("net: Retrying")
package logger

import (
	"sync"
	"time"
)

var (
	global *Logger
)

const (
	defaultTimeFormat = "[2006-01-02 15:04:05]"
)

// LoggerOption configures a Logger. Options are created by the Set*
// functions and applied by New or Init.
type LoggerOption func(*Logger)

// Logger dispatches log records to a set of handlers. It holds formatting
// settings used by Record and the bookkeeping for rate-limited entries. A
// Logger is safe for concurrent use.
type Logger struct {
	showIcons    bool
	timeFormat   string
	maxLimit     time.Duration
	limits       map[string]time.Time
	limitsLock   sync.Mutex
	limitsClean  time.Time
	handlersLock sync.Mutex
	handlers     []func(*Record)
}

// Panic logs a message at PanicLevel with no fields. The message is formed
// from args in the manner of fmt.Sprint. It does not call the built-in panic.
func (l *Logger) Panic(args ...interface{}) {
	entry := &Entry{
		logger: l,
	}
	entry.Panic(args...)
}

// Crit logs a message at CritLevel with no fields. The message is formed
// from args in the manner of fmt.Sprint.
func (l *Logger) Crit(args ...interface{}) {
	entry := &Entry{
		logger: l,
	}
	entry.Crit(args...)
}

// Error logs a message at ErrorLevel with no fields. The message is formed
// from args in the manner of fmt.Sprint.
func (l *Logger) Error(args ...interface{}) {
	entry := &Entry{
		logger: l,
	}
	entry.Error(args...)
}

// Warn logs a message at WarnLevel with no fields. The message is formed
// from args in the manner of fmt.Sprint.
func (l *Logger) Warn(args ...interface{}) {
	entry := &Entry{
		logger: l,
	}
	entry.Warn(args...)
}

// Info logs a message at InfoLevel with no fields. The message is formed
// from args in the manner of fmt.Sprint.
func (l *Logger) Info(args ...interface{}) {
	entry := &Entry{
		logger: l,
	}
	entry.Info(args...)
}

// Debug logs a message at DebugLevel with no fields. The message is formed
// from args in the manner of fmt.Sprint.
func (l *Logger) Debug(args ...interface{}) {
	entry := &Entry{
		logger: l,
	}
	entry.Debug(args...)
}

// Trace logs a message at TraceLevel with no fields. The message is formed
// from args in the manner of fmt.Sprint.
func (l *Logger) Trace(args ...interface{}) {
	entry := &Entry{
		logger: l,
	}
	entry.Trace(args...)
}

// WithFields returns an Entry carrying the given structured fields. Call a
// level method on the returned Entry to emit it. The keys "error" and
// "error_data" are treated specially by Record formatting; see Fields.
func (l *Logger) WithFields(fields Fields) *Entry {
	return &Entry{
		logger: l,
		data:   fields,
	}
}

// AddHandler registers hand to be called synchronously with every Record
// emitted by the logger, in registration order. Handlers are invoked on the
// goroutine that logged the message and should return promptly.
func (l *Logger) AddHandler(hand func(*Record)) {
	l.handlersLock.Lock()
	defer l.handlersLock.Unlock()
	l.handlers = append(l.handlers, hand)
}

// cleanLimits removes rate-limit entries whose last emission is older than
// maxLimit and records the time of the cleanup.
func (l *Logger) cleanLimits() {
	l.limitsLock.Lock()
	defer l.limitsLock.Unlock()

	now := time.Now()
	l.limitsClean = now

	for token, timestamp := range l.limits {
		if now.Sub(timestamp) > l.maxLimit {
			delete(l.limits, token)
		}
	}
}

// Init applies opts to the package-level global logger. It may be called
// more than once; each call applies only the options given.
func Init(opts ...LoggerOption) {
	for _, opt := range opts {
		opt(global)
	}
}

// New returns a new Logger with the given options applied. By default icons
// are shown, the time format is "[2006-01-02 15:04:05]" and the maximum
// rate-limit duration is one hour. The returned logger has no handlers and
// discards all records until one is added with AddHandler.
func New(opts ...LoggerOption) *Logger {
	logr := &Logger{
		showIcons:    true,
		timeFormat:   defaultTimeFormat,
		maxLimit:     1 * time.Hour,
		limits:       map[string]time.Time{},
		limitsLock:   sync.Mutex{},
		limitsClean:  time.Now(),
		handlersLock: sync.Mutex{},
		handlers:     []func(*Record){},
	}

	for _, opt := range opts {
		opt(logr)
	}

	return logr
}

// Panic logs a message at PanicLevel using the global logger. See
// Logger.Panic.
func Panic(args ...interface{}) {
	global.Panic(args...)
}

// Crit logs a message at CritLevel using the global logger. See
// Logger.Crit.
func Crit(args ...interface{}) {
	global.Crit(args...)
}

// Error logs a message at ErrorLevel using the global logger. See
// Logger.Error.
func Error(args ...interface{}) {
	global.Error(args...)
}

// Warn logs a message at WarnLevel using the global logger. See
// Logger.Warn.
func Warn(args ...interface{}) {
	global.Warn(args...)
}

// Info logs a message at InfoLevel using the global logger. See
// Logger.Info.
func Info(args ...interface{}) {
	global.Info(args...)
}

// Debug logs a message at DebugLevel using the global logger. See
// Logger.Debug.
func Debug(args ...interface{}) {
	global.Debug(args...)
}

// Trace logs a message at TraceLevel using the global logger. See
// Logger.Trace.
func Trace(args ...interface{}) {
	global.Trace(args...)
}

// WithFields returns an Entry on the global logger carrying the given
// fields. See Logger.WithFields.
func WithFields(fields Fields) *Entry {
	return global.WithFields(fields)
}

// AddHandler registers a handler on the global logger. See
// Logger.AddHandler.
func AddHandler(hand func(*Record)) {
	global.AddHandler(hand)
}

// SetTimeFormat returns an option that sets the layout, in time.Format
// syntax, used to render a record's timestamp. The default is
// "[2006-01-02 15:04:05]".
func SetTimeFormat(format string) LoggerOption {
	return func(l *Logger) {
		l.timeFormat = format
	}
}

// SetMaxLimit returns an option that sets the maximum rate-limit duration
// the logger supports. It is not a limit on log size or retention.
//
// When an Entry is emitted with Entry.Limit, the logger records the time the
// message was last emitted so that repeats within the limit window can be
// dropped. These records are garbage-collected periodically: whenever more
// than dur has elapsed since the last cleanup, any record older than dur is
// discarded. dur must therefore be at least as long as the longest duration
// ever passed to Entry.Limit, otherwise a rate-limit record may be discarded
// before its window has expired and the message will be emitted again early.
// Larger values only increase how long stale records are retained in memory.
//
// The default is one hour. The option has no observable effect if
// Entry.Limit is never used.
func SetMaxLimit(dur time.Duration) LoggerOption {
	return func(l *Logger) {
		l.maxLimit = dur
	}
}

// SetIcons returns an option that controls whether formatted records include
// the "▶" message marker and "◆" field separators. Icons are shown by
// default.
func SetIcons(show bool) LoggerOption {
	return func(l *Logger) {
		l.showIcons = show
	}
}

func init() {
	logr := &Logger{
		showIcons:    true,
		timeFormat:   defaultTimeFormat,
		maxLimit:     1 * time.Hour,
		limits:       map[string]time.Time{},
		limitsLock:   sync.Mutex{},
		limitsClean:  time.Now(),
		handlersLock: sync.Mutex{},
		handlers:     []func(*Record){},
	}

	global = logr
}
