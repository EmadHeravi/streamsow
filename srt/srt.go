package srt

import "github.com/haivision/srtgo"

// Mode wraps the srtgo socket mode.
type Mode int

const (
	ModeCaller   Mode = Mode(srtgo.ModeCaller)
	ModeListener Mode = Mode(srtgo.ModeListener)
	ModeFailure  Mode = Mode(srtgo.ModeFailure)
)

// LogLevel is a thin alias of srtgo.SrtLogLevel.
type LogLevel = srtgo.SrtLogLevel

const (
	LogLevelCrit   LogLevel = srtgo.SrtLogLevelCrit
	LogLevelErr    LogLevel = srtgo.SrtLogLevelErr
	LogLevelWarn   LogLevel = srtgo.SrtLogLevelWarning
	LogLevelNotice LogLevel = srtgo.SrtLogLevelNotice
	LogLevelInfo   LogLevel = srtgo.SrtLogLevelInfo
	LogLevelDebug  LogLevel = srtgo.SrtLogLevelDebug
)

// LogHandlerFunc matches the callback signature used for SRT logging.
type LogHandlerFunc func(level LogLevel, file string, line int, area, message string)

// Init initializes the underlying SRT library.
func Init() {
	srtgo.InitSRT()
}

// SetLogLevel sets the global SRT log level.
func SetLogLevel(level LogLevel) {
	srtgo.SrtSetLogLevel(level)
}

// SetLogHandler sets the SRT log callback handler.
func SetLogHandler(cb LogHandlerFunc) {
	if cb == nil {
		return
	}
	srtgo.SrtSetLogHandler(func(level srtgo.SrtLogLevel, file string, line int, area, message string) {
		cb(LogLevel(level), file, line, area, message)
	})
}
