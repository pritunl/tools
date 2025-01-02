package logger

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
)

var (
	timeFormat  = "[2006-01-02 15:04:05]"
	levelFormat = "[%s]"
	showIcons   = true
)

type LoggerOption func()

func SetTimeFormat(format string) LoggerOption {
	return func() {
		timeFormat = format
	}
}

func SetLevelFormat(format string) LoggerOption {
	return func() {
		levelFormat = format
	}
}

func SetIcons(show bool) LoggerOption {
	return func() {
		showIcons = show
	}
}

type Fields map[string]interface{}

type Entry struct {
	Level   string
	Message string
	Time    time.Time
	Data    Fields
}

func (e *Entry) Debug(args ...interface{}) {
	e.log(DebugLevel, args...)
}

func (e *Entry) Info(args ...interface{}) {
	e.log(InfoLevel, args...)
}

func (e *Entry) Warn(args ...interface{}) {
	e.log(WarnLevel, args...)
}

func (e *Entry) Error(args ...interface{}) {
	e.log(ErrorLevel, args...)
}

func (e *Entry) log(level string, args ...interface{}) {
	e.Level = level
	e.Message = fmt.Sprint(args...)
	e.Time = time.Now()
	e.output()
}

func (e *Entry) output() {
	var msg string
	if timeFormat != "" {
		msg += e.Time.Format(timeFormat)
	}
	if levelFormat != "" {
		msg += fmt.Sprintf(levelFormat, strings.ToUpper(e.Level))
	}
	if msg != "" {
		msg += " "
	}
	if showIcons {
		msg += "▶ "
	}
	msg += e.Message

	keys := []string{}

	var errStr string
	for key, val := range e.Data {
		if key == "error" {
			errStr = fmt.Sprintf("%s", val)
			continue
		}

		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		if showIcons {
			msg += fmt.Sprintf(" ◆ %s=%v", key,
				fmt.Sprintf("%#v", e.Data[key]))
		} else {
			msg += fmt.Sprintf(" %s=%v", key,
				fmt.Sprintf("%#v", e.Data[key]))
		}
	}

	if errStr != "" {
		msg += "\n" + errStr
	}

	if string(msg[len(msg)-1]) != "\n" {
		msg += "\n"
	}

	fmt.Print(msg)
}

func WithFields(fields Fields) *Entry {
	return &Entry{
		Data: fields,
	}
}

func Debug(args ...interface{}) {
	entry := &Entry{}
	entry.Debug(args...)
}

func Info(args ...interface{}) {
	entry := &Entry{}
	entry.Info(args...)
}

func Warn(args ...interface{}) {
	entry := &Entry{}
	entry.Warn(args...)
}

func Error(args ...interface{}) {
	entry := &Entry{}
	entry.Error(args...)
}

func Init(opts ...LoggerOption) {
	for _, opt := range opts {
		opt()
	}
}
