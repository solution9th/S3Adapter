package zlog

import (
	"io"
	"time"

	"github.com/rs/zerolog"
)

var zlog zerolog.Logger

var (
	TimeFieldFormat = time.RFC3339

	TimeFormatUnixNano = "2006-01-02 15:04:05.999999999"

	NoColor = false
)

func newWriter(w io.Writer) *zerolog.ConsoleWriter {

	// ConsoleWriter parses the JSON input and writes it in an (optionally) colorized, human-friendly format to Out.
	return &zerolog.ConsoleWriter{
		Out:        w,
		TimeFormat: TimeFormatUnixNano,
		NoColor:    NoColor,
	}
}

func NewBasicLog(w io.Writer) {
	output := newWriter(w)

	zlog = newLog(output).Logger()
}

func NewJSONLog(w io.Writer) {
	zlog = newLog(w).Logger()
}

func newLog(w io.Writer) zerolog.Context {
	zerolog.TimeFieldFormat = TimeFormatUnixNano

	return zerolog.New(w).With().Timestamp().CallerWithSkipFrameCount(2)
}

func ZDebug() *zerolog.Event {
	return zlog.Debug()
}

func ZInfo() *zerolog.Event {
	return zlog.Info()
}

func ZWarn() *zerolog.Event {
	return zlog.Warn()
}

func ZError() *zerolog.Event {
	return zlog.Error()
}

func ZFatal() *zerolog.Event {
	return zlog.Fatal()
}

// Debugf debug format
func Debugf(format string, v ...interface{}) {
	ZDebug().Msgf(format, v...)
}

// // Infof info format
// func Infof(format string, v ...interface{}) {
// 	ZInfo().Msgf(format, v...)
// }

// // Warnf warn format
// func Warnf(format string, v ...interface{}) {
// 	ZWarn().Msgf(format, v...)
// }

// // Errorf error format
// func Errorf(format string, v ...interface{}) {
// 	ZError().Msgf(format, v...)
// }

// // Fatalf fatalf
// func Fatalf(format string, v ...interface{}) {
// 	ZFatal().Msgf(format, v...)
// }
