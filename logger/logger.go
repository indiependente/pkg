package logger

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogLevel represents the logging level.
type LogLevel int

const (
	// DEBUG level logging
	DEBUG LogLevel = iota
	// INFO level logging
	INFO
	// WARNING level logging
	WARNING
	// ERROR level logging
	ERROR
	// FATAL level logging
	FATAL
	// PANIC level logging
	PANIC
	// DISABLED level logging
	DISABLED
)

// LogKey is the type each key that appears in the log should be.
type LogKey string

// String returns the string representation of the LogKey
func (lk LogKey) String() string {
	return string(lk)
}

const (
	bytesWrittenKey LogKey = "bytes_written"
	callerKey       LogKey = "caller"
	durationKey     LogKey = "duration"
	eventKey        LogKey = "event"
	hostKey         LogKey = "host"
	methodKey       LogKey = "method_key"
	remoteAddrKey   LogKey = "remote_addr_key"
	requestIDKey    LogKey = "request_id"
	serviceKey      LogKey = "service"
	signalKey       LogKey = "signal"
	statusCodeKey   LogKey = "status_code"
	uriKey          LogKey = "uri_key"
	userAgentKey    LogKey = "user_agent_key"
)

// Logger defines the behavior of the logger.
// Exposes a function for each loggable field which maps to a LogKey.
// The functions invocations can be chained and terminated by one of the levelled function calls (Fatal, Error, Warn, Info).
type Logger interface {
	BytesWritten(int) Logger
	Duration(time.Duration) Logger
	Host(string) Logger
	Method(string) Logger
	Event(string) Logger
	RequestID(string) Logger
	RemoteAddr(string) Logger
	StatusCode(int) Logger
	Signal(fmt.Stringer) Logger
	URI(string) Logger
	UserAgent(string) Logger

	// These are the last functions that should be called on a log chain.
	// These will execute and log all the information
	Panic(msg string)
	Fatal(msg string, err error)
	Error(msg string, err error)
	Warn(msg string)
	Info(msg string)
	Debug(msg string)
}

// compile time interface check.
var _ Logger = &FastLogger{}

// FastLogger implements the LogChainer interface and relies on http://github.com/rs/zerolog.
type FastLogger struct {
	lggr zerolog.Logger
}

// BytesWritten instructs the logger to log the bytes written.
func (l *FastLogger) BytesWritten(bw int) Logger {
	lcopy := *l
	lcopy.lggr = l.lggr.With().Int(bytesWrittenKey.String(), bw).Logger()
	return &lcopy
}

// Duration instructs the logger to log the duration.
func (l *FastLogger) Duration(d time.Duration) Logger {
	lcopy := *l
	lcopy.lggr = l.lggr.With().Dur(durationKey.String(), d).Logger()
	return &lcopy
}

// Host instructs the logger to log the host.
func (l *FastLogger) Host(h string) Logger {
	lcopy := *l
	lcopy.lggr = l.lggr.With().Str(hostKey.String(), h).Logger()
	return &lcopy
}

// UserAgent instructs the logger to log the user agent.
func (l *FastLogger) UserAgent(ua string) Logger {
	lcopy := *l
	lcopy.lggr = l.lggr.With().Str(userAgentKey.String(), ua).Logger()
	return &lcopy
}

// Method instructs the logger to log the method.
func (l *FastLogger) Method(m string) Logger {
	lcopy := *l
	lcopy.lggr = l.lggr.With().Str(methodKey.String(), m).Logger()
	return &lcopy
}

// Event instructs the logger to log the event.
func (l *FastLogger) Event(e string) Logger {
	lcopy := *l
	lcopy.lggr = l.lggr.With().Str(eventKey.String(), e).Logger()
	return &lcopy
}

// RequestID instructs the logger to log the request ID.
func (l *FastLogger) RequestID(id string) Logger {
	lcopy := *l
	lcopy.lggr = l.lggr.With().Str(requestIDKey.String(), id).Logger()
	return &lcopy
}

// RemoteAddr instructs the logger to log the remote address.
func (l *FastLogger) RemoteAddr(addr string) Logger {
	lcopy := *l
	lcopy.lggr = l.lggr.With().Str(remoteAddrKey.String(), addr).Logger()
	return &lcopy
}

// StatusCode instructs the logger to log the status code.
func (l *FastLogger) StatusCode(sc int) Logger {
	lcopy := *l
	lcopy.lggr = l.lggr.With().Int(statusCodeKey.String(), sc).Logger()
	return &lcopy
}

// Signal instructs the logger to log the signal.
func (l *FastLogger) Signal(sig fmt.Stringer) Logger {
	lcopy := *l
	lcopy.lggr = l.lggr.With().Str(signalKey.String(), sig.String()).Logger()
	return &lcopy
}

// URI instructs the logger to log the URI.
func (l *FastLogger) URI(uri string) Logger {
	lcopy := *l
	lcopy.lggr = l.lggr.With().Str(uriKey.String(), uri).Logger()
	return &lcopy
}

// Panic logs the message at panic level.
// It stops the ordinary flow of a goroutine.
// The log payload will contain everything else the logger has been instructed to log.
func (l *FastLogger) Panic(msg string) {
	l.lggr.Panic().Str(callerKey.String(), getCallerFunctionName()).Msg(msg)
}

// Fatal logs the message and the error at fatal level.
// It after exits with os.Exit(1).
// The log payload will contain everything else the logger has been instructed to log.
func (l *FastLogger) Fatal(msg string, err error) {
	l.lggr.Fatal().AnErr("error", err).Str(callerKey.String(), getCallerFunctionName()).Msg(msg)
}

// Error logs the message and the error at error level.
// The log payload will contain everything else the logger has been instructed to log.
func (l *FastLogger) Error(msg string, err error) {
	l.lggr.Error().AnErr("error", err).Str(callerKey.String(), getCallerFunctionName()).Msg(msg)
}

// Warn logs the message at warning level.
// The log payload will contain everything else the logger has been instructed to log.
func (l *FastLogger) Warn(msg string) {
	l.lggr.Warn().Str(callerKey.String(), getCallerFunctionName()).Msg(msg)
}

// Info logs the message at info level.
// The log payload will contain everything else the logger has been instructed to log.
func (l *FastLogger) Info(msg string) {
	l.lggr.Info().Str(callerKey.String(), getCallerFunctionName()).Msg(msg)
}

// Debug logs the message at debug level.
// The log payload will contain everything else the logger has been instructed to log.
func (l *FastLogger) Debug(msg string) {
	l.lggr.Debug().Str(callerKey.String(), getCallerFunctionName()).Msg(msg)
}

func getCallerFunctionName() string {
	// Skip GetCallerFunctionName and the function to get the caller of
	return getFrame(2).Function
}

func getFrame(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}

// GetLogger returns a pointer to a Logger that logs from logLevel and above.
// The logger is instructed to include in each log message the name of the service received in input.
func GetLogger(service string, logLevel LogLevel) *FastLogger {
	switch logLevel {
	case DEBUG:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case INFO:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case WARNING:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case ERROR:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case FATAL:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case PANIC:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case DISABLED:
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}
	return &FastLogger{
		lggr: log.With().Str(serviceKey.String(), service).Logger(),
	}
}

// GetConsoleLogger returns a pointer to a Logger that logs from logLevel and above to standard output in colorised human readable format.
// The logger is instructed to include in each log message the name of the service received in input.
func GetConsoleLogger(service string, logLevel LogLevel) *FastLogger {
	switch logLevel {
	case DEBUG:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case INFO:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case WARNING:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case ERROR:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case FATAL:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case PANIC:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case DISABLED:
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}
	return &FastLogger{
		lggr: log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Str(serviceKey.String(), service).Logger(),
	}
}

// GetLoggerString - alternative Logger constructor that returns a pointer to a Logger based on a string defining
// a log level.
// The default value is INFO.
func GetLoggerString(service string, logLevel string) *FastLogger {
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "WARNING":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "FATAL":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "PANIC":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "DISABLED":
		zerolog.SetGlobalLevel(zerolog.Disabled)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return &FastLogger{
		lggr: log.With().Str(serviceKey.String(), service).Logger(),
	}
}

// ParseLogLevel parses the input string and returns the respective log level.
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARNING":
		return WARNING
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	case "PANIC":
		return PANIC
	case "DISABLED":
		return DISABLED
	}
	return INFO
}
