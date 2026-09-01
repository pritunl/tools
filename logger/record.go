package logger

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/pritunl/tools/colorize"
	"github.com/pritunl/tools/errortypes"
)

var (
	blueArrow    = colorize.String("▶", colorize.BlueBold, colorize.None)
	whiteDiamond = colorize.String("◆", colorize.WhiteBold, colorize.None)
)

// Record is a single emitted log message as delivered to handlers. It
// carries the level, message, timestamp and structured fields, and can
// render itself as a plain or ANSI-colored line using the formatting
// settings of the Logger that produced it. Formatted output is cached after
// the first call to String or StringColor.
type Record struct {
	// Level is the severity the record was logged at.
	Level Level
	// Message is the log message text.
	Message string
	// Time is when the record was emitted.
	Time time.Time
	// Data holds the structured fields attached with WithFields. It may be
	// nil.
	Data           Fields
	logger         *Logger
	formattedPlain string
	formattedColor string
}

// Fields is the set of structured key/value pairs attached to a log entry.
//
// Two keys are handled specially when a Record is formatted. A value under
// "error" is rendered with %s on its own line after the message rather than
// as a key=value field, so that multi-line errors with stack traces remain
// readable. A value under "error_data" of type *errortypes.ErrorData is
// expanded into the fields "error_key" and "error_msg" and is otherwise
// omitted. All remaining fields are rendered as key=value pairs sorted by
// key, with values formatted using %#v.
type Fields map[string]interface{}

// String returns the record formatted as a single plain-text line
// terminated by a newline. The result is cached.
func (r *Record) String() string {
	if r.formattedPlain == "" {
		r.formatPlain()
	}
	return r.formattedPlain
}

// StringColor returns the record formatted as a single line with ANSI color
// escape sequences, terminated by a newline. The result is cached.
func (r *Record) StringColor() string {
	if r.formattedColor == "" {
		r.formatColor()
	}
	return r.formattedColor
}

// formatLevelPlain returns the bracketed level tag for plain output.
func (r *Record) formatLevelPlain() string {
	switch r.Level {
	case PanicLevel:
		return "[PANC]"
	case CritLevel:
		return "[CRIT]"
	case ErrorLevel:
		return "[ERRO]"
	case WarnLevel:
		return "[WARN]"
	case InfoLevel:
		return "[INFO]"
	case DebugLevel:
		return "[DEBG]"
	case TraceLevel:
		return "[DEBG]"
	}

	return ""
}

// formatLevelColor returns the bracketed level tag for color output, with a
// background color chosen by level.
func (r *Record) formatLevelColor() string {
	var colorBg colorize.Color

	var str string

	switch r.Level {
	case PanicLevel:
		colorBg = colorize.PurpleBg
		str = "[PANC]"
	case CritLevel:
		colorBg = colorize.PurpleBg
		str = "[CRIT]"
	case ErrorLevel:
		colorBg = colorize.RedBg
		str = "[ERRO]"
	case WarnLevel:
		colorBg = colorize.YellowBg
		str = "[WARN]"
	case InfoLevel:
		colorBg = colorize.CyanBg
		str = "[INFO]"
	case DebugLevel:
		colorBg = colorize.GrayBg
		str = "[DEBG]"
	case TraceLevel:
		colorBg = colorize.BlackBg
		str = "[TRCE]"
	default:
		colorBg = colorize.BlackBg
	}

	return colorize.String(str, colorize.WhiteBold, colorBg)
}

// formatPlain renders the record without color and stores the result in
// formattedPlain.
func (r *Record) formatPlain() {
	var msg string
	msg += r.Time.Format(r.logger.timeFormat)
	msg += r.formatLevelPlain()
	msg += " "
	if r.logger.showIcons {
		msg += "▶ "
	}
	msg += r.Message

	keys := []string{}

	var errStr string
	var errDataKey string
	var errDataMsg string
	for key, val := range r.Data {
		if key == "error" {
			errStr = fmt.Sprintf("%s", val)
			continue
		} else if key == "error_data" {
			if val != nil && !reflect.ValueOf(val).IsNil() {
				if errData, ok := val.(*errortypes.ErrorData); ok {
					errDataKey = errData.Error
					errDataMsg = errData.Message
				}
			}
			continue
		}
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		if r.logger.showIcons {
			msg += fmt.Sprintf(" ◆ %s=%v", key,
				fmt.Sprintf("%#v", r.Data[key]))
		} else {
			msg += fmt.Sprintf(" %s=%v", key,
				fmt.Sprintf("%#v", r.Data[key]))
		}
	}

	if errDataKey != "" && errDataMsg != "" {
		if r.logger.showIcons {
			msg += fmt.Sprintf(" ◆ error_key=%v",
				fmt.Sprintf("%#v", errDataKey))
			msg += fmt.Sprintf(" ◆ error_msg=%v",
				fmt.Sprintf("%#v", errDataMsg))
		} else {
			msg += fmt.Sprintf(" error_key=%v",
				fmt.Sprintf("%#v", errDataKey))
			msg += fmt.Sprintf(" error_msg=%v",
				fmt.Sprintf("%#v", errDataMsg))
		}
	}

	if errStr != "" {
		msg += "\n" + errStr
	}

	if string(msg[len(msg)-1]) != "\n" {
		msg += "\n"
	}

	r.formattedPlain = msg
}

// formatColor renders the record with ANSI colors and stores the result in
// formattedColor.
func (r *Record) formatColor() {
	var msg string
	msg += colorize.String(
		r.Time.Format(r.logger.timeFormat),
		colorize.Bold,
		colorize.None,
	)
	msg += r.formatLevelColor()
	msg += " "
	if r.logger.showIcons {
		msg += blueArrow + " "
	}
	msg += r.Message

	keys := []string{}

	var errStr string
	var errDataKey string
	var errDataMsg string
	for key, val := range r.Data {
		if key == "error" {
			errStr = fmt.Sprintf("%s", val)
			continue
		} else if key == "error_data" {
			if val != nil && !reflect.ValueOf(val).IsNil() {
				if errData, ok := val.(*errortypes.ErrorData); ok {
					errDataKey = errData.Error
					errDataMsg = errData.Message
				}
			}
			continue
		}
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		if r.logger.showIcons {
			msg += fmt.Sprintf(
				" %s %s=%v",
				whiteDiamond,
				colorize.String(key, colorize.CyanBold, colorize.None),
				colorize.String(fmt.Sprintf("%#v", r.Data[key]),
					colorize.GreenBold, colorize.None),
			)
		} else {
			msg += fmt.Sprintf(
				" %s=%v",
				colorize.String(key, colorize.CyanBold, colorize.None),
				colorize.String(fmt.Sprintf("%#v", r.Data[key]),
					colorize.GreenBold, colorize.None),
			)
		}
	}

	if errDataKey != "" && errDataMsg != "" {
		if r.logger.showIcons {
			msg += fmt.Sprintf(
				" %s %s=%v",
				whiteDiamond,
				colorize.String("error_key", colorize.CyanBold, colorize.None),
				colorize.String(fmt.Sprintf("%#v", errDataKey),
					colorize.GreenBold, colorize.None),
			)
			msg += fmt.Sprintf(
				" %s %s=%v",
				whiteDiamond,
				colorize.String("error_msg", colorize.CyanBold, colorize.None),
				colorize.String(fmt.Sprintf("%#v", errDataMsg),
					colorize.GreenBold, colorize.None),
			)
		} else {
			msg += fmt.Sprintf(
				" %s=%v",
				colorize.String("error_key", colorize.CyanBold, colorize.None),
				colorize.String(fmt.Sprintf("%#v", errDataKey),
					colorize.GreenBold, colorize.None),
			)
			msg += fmt.Sprintf(
				" %s=%v",
				colorize.String("error_msg", colorize.CyanBold, colorize.None),
				colorize.String(fmt.Sprintf("%#v", errDataMsg),
					colorize.GreenBold, colorize.None),
			)
		}
	}

	if errStr != "" {
		msg += "\n" + colorize.String(errStr, colorize.Red, colorize.None)
	}

	if string(msg[len(msg)-1]) != "\n" {
		msg += "\n"
	}

	r.formattedColor = msg
}
