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

type LoggerOption func(*Logger)

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

func (l *Logger) Panic(args ...interface{}) {
	entry := &Entry{
		logger: l,
	}
	entry.Panic(args...)
}

func (l *Logger) Crit(args ...interface{}) {
	entry := &Entry{
		logger: l,
	}
	entry.Crit(args...)
}

func (l *Logger) Error(args ...interface{}) {
	entry := &Entry{
		logger: l,
	}
	entry.Error(args...)
}

func (l *Logger) Warn(args ...interface{}) {
	entry := &Entry{
		logger: l,
	}
	entry.Warn(args...)
}

func (l *Logger) Info(args ...interface{}) {
	entry := &Entry{
		logger: l,
	}
	entry.Info(args...)
}

func (l *Logger) Debug(args ...interface{}) {
	entry := &Entry{
		logger: l,
	}
	entry.Debug(args...)
}

func (l *Logger) Trace(args ...interface{}) {
	entry := &Entry{
		logger: l,
	}
	entry.Trace(args...)
}

func (l *Logger) WithFields(fields Fields) *Entry {
	return &Entry{
		logger: l,
		data:   fields,
	}
}

func (l *Logger) AddHandler(hand func(*Record)) {
	l.handlersLock.Lock()
	defer l.handlersLock.Unlock()
	l.handlers = append(l.handlers, hand)
}

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

func Init(opts ...LoggerOption) {
	for _, opt := range opts {
		opt(global)
	}
}

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

func Panic(args ...interface{}) {
	global.Panic(args...)
}

func Crit(args ...interface{}) {
	global.Crit(args...)
}

func Error(args ...interface{}) {
	global.Error(args...)
}

func Warn(args ...interface{}) {
	global.Warn(args...)
}

func Info(args ...interface{}) {
	global.Info(args...)
}

func Debug(args ...interface{}) {
	global.Debug(args...)
}

func Trace(args ...interface{}) {
	global.Trace(args...)
}

func WithFields(fields Fields) *Entry {
	return global.WithFields(fields)
}

func AddHandler(hand func(*Record)) {
	global.AddHandler(hand)
}

func SetTimeFormat(format string) LoggerOption {
	return func(l *Logger) {
		l.timeFormat = format
	}
}

func SetMaxLimit(dur time.Duration) LoggerOption {
	return func(l *Logger) {
		l.maxLimit = dur
	}
}

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
