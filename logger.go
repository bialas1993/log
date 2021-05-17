// Package log offers simple cross platform logging for Windows and Linux.
// Available logging endpoints are event log (Windows), syslog (Linux), and
// an io.Writer.
package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type Level uint8

// Output levels.
const (
	LevelFatal Level = iota
	LevelError
	LevelWaring
	LevelInfo
	LevelDebug
	LevelDefault = LevelInfo
)

// Severity tags.
const (
	// Ltest = 0 //check iota increments
	Ldate         = 1 << iota // the date in the local time zone: 2009/01/23
	Ltime                     // the time in the local time zone: 01:23:23
	Lmicroseconds             // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                 // full file name and line number: /a/b/c/d.go:23
	Lshortfile                // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                      // if Ldate or Ltime is set, use UTC rather than the local time zone
	Lmsgprefix                // move the "prefix" from the beginning of the line to before the message
	LstdFlags     = Ldate | Ltime
	Ldisable      = 0

	tagDebug   = "DEBUG: "
	tagInfo    = "INFO : "
	tagWarning = "WARN : "
	tagError   = "ERROR: "
	tagFatal   = "FATAL: "

	initText = "ERROR: Logging before logger.Init.\n"
)

var (
	logLock       sync.Mutex
	defaultLogger *logger
	levelMap      = map[Level]string{
		LevelFatal:  "fatal",
		LevelError:  "error",
		LevelWaring: "warning",
		LevelInfo:   "info",
		LevelDebug:  "debug",
	}
)

// LogFields for add context information
type LogFields map[string]interface{}

// A logger represents an active logging object. Multiple loggers can be used
// simultaneously even if they are using the same same writers.
type logger struct {
	debugLog    *log.Logger
	infoLog     *log.Logger
	warningLog  *log.Logger
	errorLog    *log.Logger
	fatalLog    *log.Logger
	formatter   Formatter
	closers     []io.Closer
	initialized bool
	level       Level
	flags       int
	fields      LogFields
}

// LogOption modify logger instance
type LogOption func(*logger)

// initialize resets defaultLogger.  Which allows tests to reset environment.
func initialize() {
	defaultLogger = &logger{
		debugLog:   log.New(os.Stderr, initText+tagDebug, Ldate|Lmicroseconds|Lshortfile),
		infoLog:    log.New(os.Stderr, initText+tagInfo, Ldate|Lmicroseconds|Lshortfile),
		warningLog: log.New(os.Stderr, initText+tagWarning, Ldate|Lmicroseconds|Lshortfile),
		errorLog:   log.New(os.Stderr, initText+tagError, Ldate|Lmicroseconds|Lshortfile),
		fatalLog:   log.New(os.Stderr, initText+tagFatal, Ldate|Lmicroseconds|Lshortfile),
		formatter:  StdFormatter{},
		fields:     LogFields{},
		level:      LevelDefault,
		flags:      LstdFlags,
	}
}

func init() {
	initialize()
	NewStdLogger()
}

// new sets up logging and should be called before log functions.
// The first call to new populates the default logger and returns the
// generated logger, subsequent calls to Init will only return the generated
// logger.
// If the logFile passed in also satisfies io.Closer, logFile.Close will be called
// when closing the logger.
func new(name string, systemLog bool, logFile io.Writer, opts ...LogOption) *logger {
	var dl, il, wl, el io.Writer
	var syslogErr error
	dLogs, iLogs, wLogs, eLogs := []io.Writer{}, []io.Writer{}, []io.Writer{}, []io.Writer{}

	if systemLog {
		dl, il, wl, el, syslogErr = setup(name)
	}

	if logFile != nil {
		dLogs = append(dLogs, logFile)
		iLogs = append(iLogs, logFile)
		wLogs = append(wLogs, logFile)
		eLogs = append(eLogs, logFile)
	}

	if dl != nil {
		dLogs = append(dLogs, il)
	}
	if il != nil {
		iLogs = append(iLogs, il)
	}
	if wl != nil {
		wLogs = append(wLogs, wl)
	}
	if el != nil {
		eLogs = append(eLogs, el)
	}

	l := logger{
		formatter: StdFormatter{},
		flags:     LstdFlags,
		fields:    LogFields{},
		level:     LevelDefault,
	}

	for _, opt := range opts {
		opt(&l)
	}

	// Windows services don't have stdout/stderr. Writes will fail, so try them last.
	dLogs = append(dLogs, os.Stdout)
	iLogs = append(iLogs, os.Stdout)
	wLogs = append(wLogs, os.Stdout)
	eLogs = append(eLogs, os.Stderr)

	prefixDebug, prefixInfo, prefixWaring, prefixError, prefixFatal := tagDebug, tagInfo, tagWarning, tagError, tagFatal
	if l.formatter.HasSettings() {
		l.flags = Ldisable
		prefixDebug, prefixInfo, prefixWaring, prefixError, prefixFatal = "", "", "", "", ""
	}

	l.debugLog = log.New(io.MultiWriter(dLogs...), prefixDebug, l.flags)
	l.infoLog = log.New(io.MultiWriter(iLogs...), prefixInfo, l.flags)
	l.warningLog = log.New(io.MultiWriter(wLogs...), prefixWaring, l.flags)
	l.errorLog = log.New(io.MultiWriter(eLogs...), prefixError, l.flags)
	l.fatalLog = log.New(io.MultiWriter(eLogs...), prefixFatal, l.flags)

	for _, w := range []io.Writer{logFile, il, wl, el} {
		if c, ok := w.(io.Closer); ok && c != nil {
			l.closers = append(l.closers, c)
		}
	}

	l.initialized = true

	if syslogErr != nil {
		l.Error(syslogErr)
	}

	logLock.Lock()
	defer logLock.Unlock()
	if !defaultLogger.initialized {
		defaultLogger = &l
	}

	return &l
}

// NewSyslogLogger with logging to system log
func NewSyslogLogger(name string, opts ...LogOption) Logger {
	return new(name, true, nil, opts...)
}

// NewStdLogger standard console logging
func NewStdLogger(opts ...LogOption) Logger {
	return new("", false, nil, opts...)
}

// NewJsonLogger with json formatter
func NewJsonLogger(opts ...LogOption) Logger {
	return new("", false, nil, append([]LogOption{WithFormatter(JsonFormatter{})}, opts...)...)
}

// New create standard logger instance
func New(out io.Writer, opts ...LogOption) Logger {
	return new("", false, out, opts...)
}

// Close closes the default logger.
func Close() {
	defaultLogger.Close()
}

func WithFormatter(f Formatter) LogOption {
	return func(l *logger) {
		l.formatter = f
	}
}

func (l LogFields) Add(newFields LogFields) LogFields {
	if len(l) == 0 {
		return newFields
	} else if len(newFields) == 0 {
		return l
	}

	resultFields := make(LogFields, len(l)+len(newFields))
	for field, value := range l {
		resultFields[field] = value
	}

	for field, value := range newFields {
		resultFields[field] = value
	}

	return resultFields
}

func (l *logger) clear() {
	logLock.Lock()
	defer logLock.Unlock()
	l.fields = LogFields{}
}

func (l *logger) output(s Level, depth int, txt string) {
	defer l.clear()

	if l.level >= s {
		logLock.Lock()
		defer logLock.Unlock()
		switch s {
		case LevelDebug:
			l.debugLog.Output(3+depth, txt)
		case LevelInfo:
			l.infoLog.Output(3+depth, txt)
		case LevelWaring:
			l.warningLog.Output(3+depth, txt)
		case LevelError:
			l.errorLog.Output(3+depth, txt)
		case LevelFatal:
			l.fatalLog.Output(3+depth, txt)
		}
	}
}

type Logger interface {
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Warning(v ...interface{})
	Warningf(format string, v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	SetLevel(lvl Level)
	SetFlags(flag int)
	With(fields LogFields) Logger
	Close()
}

// Close closes all the underlying log writers, which will flush any cached logs.
// Any errors from closing the underlying log writers will be printed to stderr.
// Once Close is called, all future calls to the logger will panic.
func (l *logger) Close() {
	logLock.Lock()
	defer logLock.Unlock()

	if !l.initialized {
		return
	}

	for _, c := range l.closers {
		if err := c.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close log %v: %v\n", c, err)
		}
	}
}

// Debug logs with the Debug severity.
// Arguments are handled in the manner of fmt.Print.
func (l *logger) Debug(v ...interface{}) {
	l.output(LevelDebug, 0, string(l.formatter.Output(l.flags, levelMap[LevelDebug], l.fields, fmt.Sprint(v...))))
}

// Debugf logs with the Debug severity.
// Arguments are handled in the manner of fmt.Printf.
func (l *logger) Debugf(format string, v ...interface{}) {
	l.output(LevelDebug, 0, string(l.formatter.Output(l.flags, levelMap[LevelDebug], l.fields, fmt.Sprintf(format, v...))))
}

// Info logs with the Info severity.
// Arguments are handled in the manner of fmt.Print.
func (l *logger) Info(v ...interface{}) {
	l.output(LevelInfo, 0, string(l.formatter.Output(l.flags, levelMap[LevelInfo], l.fields, fmt.Sprint(v...))))
}

// Infof logs with the Info severity.
// Arguments are handled in the manner of fmt.Printf.
func (l *logger) Infof(format string, v ...interface{}) {
	l.output(LevelInfo, 0, string(l.formatter.Output(l.flags, levelMap[LevelInfo], l.fields, fmt.Sprintf(format, v...))))
}

// Warning logs with the Warning severity.
// Arguments are handled in the manner of fmt.Print.
func (l *logger) Warning(v ...interface{}) {
	l.output(LevelWaring, 0, string(l.formatter.Output(l.flags, levelMap[LevelWaring], l.fields, fmt.Sprint(v...))))
}

// Warningf logs with the Warning severity.
// Arguments are handled in the manner of fmt.Printf.
func (l *logger) Warningf(format string, v ...interface{}) {
	l.output(LevelWaring, 0, string(l.formatter.Output(l.flags, levelMap[LevelWaring], l.fields, fmt.Sprintf(format, v...))))
}

// Fatal logs with the Fatal severity, and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Print.
func (l *logger) Fatal(v ...interface{}) {
	l.output(LevelFatal, 0, string(l.formatter.Output(l.flags, levelMap[LevelFatal], l.fields, fmt.Sprint(v...))))
	l.Close()
	os.Exit(1)
}

// Fatalf logs with the Fatal severity, and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Printf.
func (l *logger) Fatalf(format string, v ...interface{}) {
	l.output(LevelFatal, 0, string(l.formatter.Output(l.flags, levelMap[LevelFatal], l.fields, fmt.Sprintf(format, v...))))
	l.Close()
	os.Exit(1)
}

// Error logs with the ERROR severity.
// Arguments are handled in the manner of fmt.Print.
func (l *logger) Error(v ...interface{}) {
	l.output(LevelError, 0, string(l.formatter.Output(l.flags, levelMap[LevelError], l.fields, fmt.Sprint(v...))))
}

// Errorf logs with the Error severity.
// Arguments are handled in the manner of fmt.Printf.
func (l *logger) Errorf(format string, v ...interface{}) {
	l.output(LevelError, 0, string(l.formatter.Output(l.flags, levelMap[LevelError], l.fields, fmt.Sprintf(format, v...))))
}

// SetLevel sets the logger verbosity level for verbose info logging.
func (l *logger) SetLevel(lvl Level) {
	l.level = lvl
}

func (l *logger) SetFlags(flag int) {
	if !l.formatter.HasSettings() {
		l.debugLog.SetFlags(flag)
		l.infoLog.SetFlags(flag)
		l.warningLog.SetFlags(flag)
		l.errorLog.SetFlags(flag)
		l.fatalLog.SetFlags(flag)
	}

	l.flags = flag
}

// With sets context fields
func (l *logger) With(fields LogFields) Logger {
	logLock.Lock()
	defer logLock.Unlock()

	l.fields = l.fields.Add(fields)

	return l
}

// SetFlags sets the output flags for the logger.
func SetFlags(flag int) {
	defaultLogger.SetFlags(flag)
}

// SetLevel sets the verbosity level for verbose info logging in the
// default logger.
func SetLevel(lvl Level) {
	defaultLogger.SetLevel(lvl)
}

// Debug uses the default logger, logs with Debug severity.
// Arguments are handled in the manner of fmt.Print.
func Debug(v ...interface{}) {
	defaultLogger.output(LevelDebug, 0, string(defaultLogger.formatter.Output(defaultLogger.flags, levelMap[LevelDebug], defaultLogger.fields, fmt.Sprint(v...))))
}

// Debugf uses the default logger, logs with Debug severity.
// Arguments are handled in the manner of fmt.Printf.
func Debugf(format string, v ...interface{}) {
	defaultLogger.output(LevelDebug, 0, string(defaultLogger.formatter.Output(defaultLogger.flags, levelMap[LevelDebug], defaultLogger.fields, fmt.Sprintf(format, v...))))
}

// Info uses the default logger and logs with the Info severity.
// Arguments are handled in the manner of fmt.Print.
func Info(v ...interface{}) {
	defaultLogger.output(LevelInfo, 0, string(defaultLogger.formatter.Output(defaultLogger.flags, levelMap[LevelInfo], defaultLogger.fields, fmt.Sprint(v...))))
}

// Infof uses the default logger and logs with the Info severity.
// Arguments are handled in the manner of fmt.Printf.
func Infof(format string, v ...interface{}) {
	defaultLogger.output(LevelInfo, 0, string(defaultLogger.formatter.Output(defaultLogger.flags, levelMap[LevelInfo], defaultLogger.fields, fmt.Sprintf(format, v...))))
}

// Warning uses the default logger and logs with the Warning severity.
// Arguments are handled in the manner of fmt.Print.
func Warning(v ...interface{}) {
	defaultLogger.output(LevelWaring, 0, string(defaultLogger.formatter.Output(defaultLogger.flags, levelMap[LevelWaring], defaultLogger.fields, fmt.Sprint(v...))))
}

// Warningf uses the default logger and logs with the Warning severity.
// Arguments are handled in the manner of fmt.Printf.
func Warningf(format string, v ...interface{}) {
	defaultLogger.output(LevelWaring, 0, string(defaultLogger.formatter.Output(defaultLogger.flags, levelMap[LevelWaring], defaultLogger.fields, fmt.Sprintf(format, v...))))
}

// Fatal uses the default logger, logs with the Fatal severity,
// and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Print.
func Fatal(v ...interface{}) {
	defaultLogger.output(LevelFatal, 0, string(defaultLogger.formatter.Output(defaultLogger.flags, levelMap[LevelFatal], defaultLogger.fields, fmt.Sprint(v...))))
	defaultLogger.Close()
	os.Exit(1)
}

// Fatalf uses the default logger, logs with the Fatal severity,
// and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Printf.
func Fatalf(format string, v ...interface{}) {
	defaultLogger.output(LevelFatal, 0, string(defaultLogger.formatter.Output(defaultLogger.flags, levelMap[LevelFatal], defaultLogger.fields, fmt.Sprintf(format, v...))))
	defaultLogger.Close()
	os.Exit(1)
}

// Error uses the default logger and logs with the Error severity.
// Arguments are handled in the manner of fmt.Print.
func Error(v ...interface{}) {
	defaultLogger.output(LevelError, 0, string(defaultLogger.formatter.Output(defaultLogger.flags, levelMap[LevelError], defaultLogger.fields, fmt.Sprint(v...))))
}

// Errorf uses the default logger and logs with the Error severity.
// Arguments are handled in the manner of fmt.Printf.
func Errorf(format string, v ...interface{}) {
	defaultLogger.output(LevelError, 0, string(defaultLogger.formatter.Output(defaultLogger.flags, levelMap[LevelError], defaultLogger.fields, fmt.Sprintf(format, v...))))
}

// With uses the default logger and store context fields for log
func With(fields LogFields) Logger {
	defaultLogger.With(fields)
	return defaultLogger
}
