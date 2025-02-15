package logger

type Level uint32

const (
	PanicLevel Level = 1
	CritLevel  Level = 2
	ErrorLevel Level = 3
	WarnLevel  Level = 4
	InfoLevel  Level = 5
	DebugLevel Level = 6
	TraceLevel Level = 7
)
