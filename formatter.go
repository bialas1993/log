package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type Formatter interface {
	// Output method should create a formatted string to display
	Output(flags int, lvl string, fields LogFields, msg string) string

	// HasFlags method should return value to decide about override log level flags
	HasFlags() bool

	// HasPrefixes method should return value to decide about override logger prefixes
	HasPrefixes() bool

	Flags() int

	Prefixes() map[Level]string
}

type StdFormatter struct{}

func (f StdFormatter) formatFields(fields LogFields) string {
	fieldsStr := ""

	keys := make([]string, len(fields))
	i := 0
	for field := range fields {
		keys[i] = field
		i++
	}

	sort.Strings(keys)

	for _, key := range keys {
		var valueStr string
		value := fields[key]

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

func (f StdFormatter) HasFlags() bool {
	return false
}

func (f StdFormatter) HasPrefixes() bool {
	return false
}

func (f StdFormatter) Flags() int {
	return 0
}

func (f StdFormatter) Prefixes() map[Level]string {
	return nil
}

func (f StdFormatter) Output(flags int, lvl string, fields LogFields, msg string) string {
	buf := bytes.NewBufferString(f.formatFields(fields))
	buf.WriteString(msg)

	return buf.String()
}

type JsonFormatter struct {
	mu sync.Mutex
}

func (f *JsonFormatter) createHeadersFields(flags int) LogFields {
	var timeBuffer bytes.Buffer
	var fileBuffer bytes.Buffer
	var file string
	var line int
	fields := LogFields{}

	t := time.Now()

	f.mu.Lock()
	defer f.mu.Unlock()
	if flags&(Lshortfile|Llongfile) != 0 {
		// Release lock while getting caller info - it's expensive.
		f.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(3)
		if !ok {
			file = "???"
			line = 0
		}
		f.mu.Lock()
	}

	if flags&(Ldate|Ltime|Lmicroseconds) != 0 {
		if flags&LUTC != 0 {
			t = t.UTC()
		}
		if flags&Ldate != 0 {
			year, month, day := t.Date()
			timeBuffer.Write(itoa(year, 4))
			timeBuffer.WriteByte('/')
			timeBuffer.Write(itoa(int(month), 2))
			timeBuffer.WriteByte('/')
			timeBuffer.Write(itoa(day, 2))
			timeBuffer.WriteByte(' ')
		}
		if flags&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			timeBuffer.Write(itoa(hour, 2))
			timeBuffer.WriteByte(':')
			timeBuffer.Write(itoa(min, 2))
			timeBuffer.WriteByte(':')
			timeBuffer.Write(itoa(sec, 2))
			if flags&Lmicroseconds != 0 {
				timeBuffer.WriteByte('.')
				timeBuffer.Write(itoa(t.Nanosecond()/1e3, 6))

			}
			timeBuffer.WriteByte(' ')
		}

		fields["time"] = strings.TrimRight(timeBuffer.String(), " ")
	}
	if flags&(Lshortfile|Llongfile) != 0 {
		if flags&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}

		fileBuffer.WriteString(file)
		fileBuffer.WriteByte(':')
		fileBuffer.Write(itoa(line, -1))

		fields["file"] = fileBuffer.String()
	}

	return fields
}

func (f JsonFormatter) formatFields(fields LogFields) string {
	b, _ := json.Marshal(fields)

	return string(b)
}

func (f JsonFormatter) Output(flags int, lvl string, fields LogFields, msg string) string {
	headersFields := f.createHeadersFields(flags)
	msgFields := LogFields{"msg": msg, "level": lvl}
	ff := f.formatFields(fields.Add(msgFields).Add(headersFields))
	buf := bytes.NewBufferString(ff)

	return buf.String()
}

func (f JsonFormatter) HasFlags() bool {
	return true
}

func (f JsonFormatter) HasPrefixes() bool {
	return true
}

func (f JsonFormatter) Flags() int {
	return Ldisable
}

func (f JsonFormatter) Prefixes() map[Level]string {
	return map[Level]string{
		LevelDebug:  "",
		LevelError:  "",
		LevelFatal:  "",
		LevelWaring: "",
		LevelInfo:   "",
	}
}

type ColorizedStdFormatter struct {
	StdFormatter
}

func (ColorizedStdFormatter) HasPrefixes() bool {
	return true
}

var (
	CLR_0 = "\x1b[30;1m"
	CLR_R = "\x1b[31;1m"
	CLR_G = "\x1b[32;1m"
	CLR_Y = "\x1b[33;1m"
	CLR_B = "\x1b[34;1m"
	CLR_M = "\x1b[35;1m"
	CLR_C = "\x1b[36;1m"
	CLR_W = "\x1b[37;1m"
	RESET = "\x1b[0m"
)

func (ColorizedStdFormatter) Prefixes() map[Level]string {
	return map[Level]string{
		LevelDebug:  CLR_W + "DEBUG: " + RESET,
		LevelPanic:  CLR_0 + "PANIC: " + RESET,
		LevelError:  CLR_R + "ERROR: " + RESET,
		LevelFatal:  CLR_R + "FATAL: " + RESET,
		LevelWaring: CLR_Y + "WARN : " + RESET,
		LevelInfo:   CLR_C + "INFO : " + RESET,
	}
}
