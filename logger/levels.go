package logger

// Level is the severity of a log record. Lower values are more severe;
// PanicLevel is the most severe and TraceLevel the least.
type Level uint32

// Log severity levels, ordered from most to least severe.
const (
	// PanicLevel is for unrecoverable failures. Logging at this level does
	// not itself panic; it only tags the record.
	PanicLevel Level = 1
	// CritLevel is for critical failures that require immediate attention.
	CritLevel Level = 2
	// ErrorLevel is for failures of an operation that the program can
	// continue past.
	ErrorLevel Level = 3
	// WarnLevel is for unexpected but non-fatal conditions.
	WarnLevel Level = 4
	// InfoLevel is for routine operational messages.
	InfoLevel Level = 5
	// DebugLevel is for verbose diagnostic messages.
	DebugLevel Level = 6
	// TraceLevel is for the most granular diagnostic messages.
	TraceLevel Level = 7
)
