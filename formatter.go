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
	// HasSettings method should return value to decide about override logger prefixes and log level flags
	HasSettings() bool
	// Output method should create a formatted string to display
	Output(flags int, lvl string, fields LogFields, msg string) string
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

func (f StdFormatter) HasSettings() bool {
	return false
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

func (f JsonFormatter) HasSettings() bool {
	return true
}

func (l LogFields) MarshalJSON() ([]byte, error) {
	var b []byte
	buf := bytes.NewBuffer(b)
	buf.WriteRune('{')

	data := [][]interface{}{}
	if v, ok := l["time"]; ok {
		data = append(data, []interface{}{"time", v})
		delete(l, "time")
	}

	if v, ok := l["level"]; ok {
		data = append(data, []interface{}{"level", v})
		delete(l, "level")
	}

	if v, ok := l["msg"]; ok {
		data = append(data, []interface{}{"msg", v})
		delete(l, "msg")
	}

	for key, val := range l {
		data = append(data, []interface{}{key, val})
	}

	for i, d := range data {
		km, err := json.Marshal(d[0])
		if err != nil {
			return nil, err
		}
		buf.Write(km)
		buf.WriteRune(':')
		vm, err := json.Marshal(d[1])
		if err != nil {
			return nil, err
		}
		buf.Write(vm)
		if i != len(data)-1 {
			buf.WriteRune(',')
		}
	}
	buf.WriteRune('}')
	return buf.Bytes(), nil
}
