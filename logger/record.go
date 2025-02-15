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

type Record struct {
	Level          Level
	Message        string
	Time           time.Time
	Data           Fields
	logger         *Logger
	formattedPlain string
	formattedColor string
}

type Fields map[string]interface{}

func (r *Record) String() string {
	if r.formattedPlain == "" {
		r.formatPlain()
	}
	return r.formattedPlain
}

func (r *Record) StringColor() string {
	if r.formattedColor == "" {
		r.formatColor()
	}
	return r.formattedColor
}

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
