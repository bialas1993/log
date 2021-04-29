// Package log offers simple cross platform logging for Windows and Linux.
// Available logging endpoints are event log (Windows), syslog (Linux), and
// an io.Writer.
package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
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
	closers     []io.Closer
	initialized bool
	level       int
	fields      LogFields
}

// LogOption modify logger instance
type LogOption func(*logger)

// Output levels.
const (
	LevelFatal int = 1 << iota
	LevelError
	LevelWaring
	LevelInfo
	LevelDebug
	LevelDefault = LevelFatal | LevelError | LevelWaring | LevelInfo
)

const (
	sFatal uint8 = iota
	sError
	sWarning
	sInfo
	sDebug
)

// Severity tags.
const (
	Ldate         = 1 << iota // the date in the local time zone: 2009/01/23
	Ltime                     // the time in the local time zone: 01:23:23
	Lmicroseconds             // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                 // full file name and line number: /a/b/c/d.go:23
	Lshortfile                // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                      // if Ldate or Ltime is set, use UTC rather than the local time zone
	Lmsgprefix                // move the "prefix" from the beginning of the line to before the message
	LstdFlags     = Ldate | Ltime

	tagDebug   = "DEBUG: "
	tagInfo    = "INFO : "
	tagWarning = "WARN : "
	tagError   = "ERROR: "
	tagFatal   = "FATAL: "
)

const (
	flags    = log.Ldate | log.Lmicroseconds | log.Lshortfile
	initText = "ERROR: Logging before logger.Init.\n"
)

var (
	logLock       sync.Mutex
	defaultLogger *logger
)

// initialize resets defaultLogger.  Which allows tests to reset environment.
func initialize() {
	defaultLogger = &logger{
		debugLog:   log.New(os.Stderr, initText+tagDebug, flags),
		infoLog:    log.New(os.Stderr, initText+tagInfo, flags),
		warningLog: log.New(os.Stderr, initText+tagWarning, flags),
		errorLog:   log.New(os.Stderr, initText+tagError, flags),
		fatalLog:   log.New(os.Stderr, initText+tagFatal, flags),
		fields:     LogFields{},
		level:      LevelDefault,
	}
}

func init() {
	initialize()
	NewStdLogger()
}

// Init sets up logging and should be called before log functions, usually in
// the caller's main(). Default log functions can be called before Init(), but log
// output will only go to stderr (along with a warning).
// The first call to Init populates the default logger and returns the
// generated logger, subsequent calls to Init will only return the generated
// logger.
// If the logFile passed in also satisfies io.Closer, logFile.Close will be called
// when closing the logger.
func new(name string, systemLog bool, logFile io.Writer, opts ...LogOption) *logger {
	var dl, il, wl, el io.Writer
	var syslogErr error
	dLogs := []io.Writer{}
	iLogs := []io.Writer{}
	wLogs := []io.Writer{}
	eLogs := []io.Writer{}

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
		fields: LogFields{},
		level:  LevelDefault,
	}

	for _, opt := range opts {
		opt(&l)
	}

	// Windows services don't have stdout/stderr. Writes will fail, so try them last.
	dLogs = append(dLogs, os.Stdout)
	iLogs = append(iLogs, os.Stdout)
	wLogs = append(wLogs, os.Stdout)
	eLogs = append(eLogs, os.Stderr)

	l.debugLog = log.New(io.MultiWriter(dLogs...), tagDebug, flags)
	l.infoLog = log.New(io.MultiWriter(iLogs...), tagInfo, flags)
	l.warningLog = log.New(io.MultiWriter(wLogs...), tagWarning, flags)
	l.errorLog = log.New(io.MultiWriter(eLogs...), tagError, flags)
	l.fatalLog = log.New(io.MultiWriter(eLogs...), tagFatal, flags)

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

func NewSyslogLogger(name string, opts ...LogOption) Logger {
	return new(name, true, nil)
}

func NewStdLogger(opts ...LogOption) Logger {
	return new("", false, nil)
}

func New(out io.Writer, opts ...LogOption) Logger {
	return new("", false, nil)
}

// Close closes the default logger.
func Close() {
	defaultLogger.Close()
}

func (l *logger) clear() {
	l.fields = LogFields{}
}

func (l *logger) formatFields() string {
	fieldsStr := ""

	keys := make([]string, len(l.fields))
	i := 0
	for field := range l.fields {
		keys[i] = field
		i++
	}

	sort.Strings(keys)

	for _, key := range keys {
		var valueStr string
		value := l.fields[key]

		if stringer, ok := value.(fmt.Stringer); ok {
			valueStr = stringer.String()
		} else {
			valueStr = fmt.Sprintf("%v", value)
		}

		if strings.Contains(valueStr, " ") {
			valueStr = `"` + valueStr + `"`
		}

		fieldsStr += key + "=" + valueStr + " "
	}

	return fieldsStr
}

func (l *logger) output(s uint8, depth int, txt string) {
	defer l.clear()

	if l.level&(1<<s) != 0 {
		buf := bytes.NewBufferString(l.formatFields())
		buf.WriteString(txt)

		logLock.Lock()
		defer logLock.Unlock()
		switch s {
		case sDebug:
			l.debugLog.Output(3+depth, buf.String())
		case sInfo:
			l.infoLog.Output(3+depth, buf.String())
		case sWarning:
			l.warningLog.Output(3+depth, buf.String())
		case sError:
			l.errorLog.Output(3+depth, buf.String())
		case sFatal:
			l.fatalLog.Output(3+depth, buf.String())
		default:
			panic(fmt.Sprintln("unrecognized severity:", s))
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
	SetLevel(lvl int)
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
	l.output(sDebug, 0, fmt.Sprint(v...))
}

// Debugf logs with the Debug severity.
// Arguments are handled in the manner of fmt.Printf.
func (l *logger) Debugf(format string, v ...interface{}) {
	l.output(sDebug, 0, fmt.Sprintf(format, v...))
}

// Info logs with the Info severity.
// Arguments are handled in the manner of fmt.Print.
func (l *logger) Info(v ...interface{}) {
	l.output(sInfo, 0, fmt.Sprint(v...))
}

// Infof logs with the Info severity.
// Arguments are handled in the manner of fmt.Printf.
func (l *logger) Infof(format string, v ...interface{}) {
	l.output(sInfo, 0, fmt.Sprintf(format, v...))
}

// Warning logs with the Warning severity.
// Arguments are handled in the manner of fmt.Print.
func (l *logger) Warning(v ...interface{}) {
	l.output(sWarning, 0, fmt.Sprint(v...))
}

// Warningf logs with the Warning severity.
// Arguments are handled in the manner of fmt.Printf.
func (l *logger) Warningf(format string, v ...interface{}) {
	l.output(sWarning, 0, fmt.Sprintf(format, v...))
}

// Fatal logs with the Fatal severity, and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Print.
func (l *logger) Fatal(v ...interface{}) {
	l.output(sFatal, 0, fmt.Sprint(v...))
	l.Close()
	os.Exit(1)
}

// Fatalf logs with the Fatal severity, and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Printf.
func (l *logger) Fatalf(format string, v ...interface{}) {
	l.output(sFatal, 0, fmt.Sprintf(format, v...))
	l.Close()
	os.Exit(1)
}

// Error logs with the ERROR severity.
// Arguments are handled in the manner of fmt.Print.
func (l *logger) Error(v ...interface{}) {
	l.output(sError, 0, fmt.Sprint(v...))
}

// Errorf logs with the Error severity.
// Arguments are handled in the manner of fmt.Printf.
func (l *logger) Errorf(format string, v ...interface{}) {
	l.output(sError, 0, fmt.Sprintf(format, v...))
}

// SetLevel sets the logger verbosity level for verbose info logging.
func (l *logger) SetLevel(lvl int) {
	l.level = lvl
}

func (l *logger) SetFlags(flag int) {
	l.debugLog.SetFlags(flag)
	l.infoLog.SetFlags(flag)
	l.warningLog.SetFlags(flag)
	l.errorLog.SetFlags(flag)
	l.fatalLog.SetFlags(flag)
}

// With sets context fields
func (l *logger) With(fields LogFields) Logger {
	l.fields = fields
	return l
}

// SetFlags sets the output flags for the logger.
func SetFlags(flag int) {
	defaultLogger.debugLog.SetFlags(flag)
	defaultLogger.infoLog.SetFlags(flag)
	defaultLogger.warningLog.SetFlags(flag)
	defaultLogger.errorLog.SetFlags(flag)
	defaultLogger.fatalLog.SetFlags(flag)
}

// SetLevel sets the verbosity level for verbose info logging in the
// default logger.
func SetLevel(lvl int) {
	defaultLogger.SetLevel(lvl)
}

// Debug uses the default logger, logs with Debug severity.
// Arguments are handled in the manner of fmt.Print.
func Debug(v ...interface{}) {
	defaultLogger.output(sDebug, 0, fmt.Sprint(v...))
}

// Debugf uses the default logger, logs with Debug severity.
// Arguments are handled in the manner of fmt.Printf.
func Debugf(format string, v ...interface{}) {
	defaultLogger.output(sDebug, 0, fmt.Sprintf(format, v...))
}

// Info uses the default logger and logs with the Info severity.
// Arguments are handled in the manner of fmt.Print.
func Info(v ...interface{}) {
	defaultLogger.output(sInfo, 0, fmt.Sprint(v...))
}

// Infof uses the default logger and logs with the Info severity.
// Arguments are handled in the manner of fmt.Printf.
func Infof(format string, v ...interface{}) {
	defaultLogger.output(sInfo, 0, fmt.Sprintf(format, v...))
}

// Warning uses the default logger and logs with the Warning severity.
// Arguments are handled in the manner of fmt.Print.
func Warning(v ...interface{}) {
	defaultLogger.output(sWarning, 0, fmt.Sprint(v...))
}

// Warningf uses the default logger and logs with the Warning severity.
// Arguments are handled in the manner of fmt.Printf.
func Warningf(format string, v ...interface{}) {
	defaultLogger.output(sWarning, 0, fmt.Sprintf(format, v...))
}

// Fatal uses the default logger, logs with the Fatal severity,
// and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Print.
func Fatal(v ...interface{}) {
	defaultLogger.output(sFatal, 0, fmt.Sprint(v...))
	defaultLogger.Close()
	os.Exit(1)
}

// Fatalf uses the default logger, logs with the Fatal severity,
// and ends with os.Exit(1).
// Arguments are handled in the manner of fmt.Printf.
func Fatalf(format string, v ...interface{}) {
	defaultLogger.output(sFatal, 0, fmt.Sprintf(format, v...))
	defaultLogger.Close()
	os.Exit(1)
}

// Error uses the default logger and logs with the Error severity.
// Arguments are handled in the manner of fmt.Print.
func Error(v ...interface{}) {
	defaultLogger.output(sError, 0, fmt.Sprint(v...))
}

// Errorf uses the default logger and logs with the Error severity.
// Arguments are handled in the manner of fmt.Printf.
func Errorf(format string, v ...interface{}) {
	defaultLogger.output(sError, 0, fmt.Sprintf(format, v...))
}

// With uses the default logger and store context fields for log
func With(fields LogFields) Logger {
	defaultLogger.fields = fields
	return defaultLogger
}
