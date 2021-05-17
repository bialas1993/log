# logger #
Logger is a simple cross platform Go logging library for Windows, Linux, FreeBSD, and
macOS, it can log to the Windows event log, Linux/macOS syslog, and an io.Writer.

## Usage ##

Set up the default logger to log the system log (event log or syslog):

```go
package main

import (
  "github.com/bialas1993/log"
)

func main() {
	log.Debug("debug")
	log.With(log.LogFields{
		"asd":   "bsd",
		"lorem": "ipsum",
		"bang":  10,
		"struct": struct {
			A string
		}{"aaaaaa"},
	}).Info("info")
}
```

## Custom Format ##

| Code                              | Example                                                  |
|-----------------------------------|----------------------------------------------------------|
| `log.SetFlags(log.Ldate)`         | ERROR: 2018/11/11 Error running Foobar: message          |
| `log.SetFlags(log.Ltime)`         | ERROR: 09:42:45 Error running Foobar: message            |
| `log.SetFlags(log.Lmicroseconds)` | ERROR: 09:42:50.776015 Error running Foobar: message     |
| `log.SetFlags(log.Llongfile)`     | ERROR: /src/main.go:31: Error running Foobar: message    |
| `log.SetFlags(log.Lshortfile)`    | ERROR: main.go:31: Error running Foobar: message         |
| `log.SetFlags(log.LUTC)`          | ERROR: Error running Foobar: message                     |
| `log.SetFlags(log.LstdFlags)`     | ERROR: 2018/11/11 09:43:12 Error running Foobar: message |


More info: https://golang.org/pkg/log/#pkg-constants