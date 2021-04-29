# logger #
Logger is a simple cross platform Go logging library for Windows, Linux, FreeBSD, and
macOS, it can log to the Windows event log, Linux/macOS syslog, and an io.Writer.

This is not an official Google product.

## Usage ##

Set up the default logger to log the system log (event log or syslog) and a
file, include a flag to turn up verbosity:

```go
import (
  "flag"
  "os"

  "github.com/bialas1993/log"
)

const logPath = "/some/location/example.log"

var verbose = flag.Bool("verbose", false, "print info level logs to stdout")

func main() {
  flag.Parse()

  lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
  if err != nil {
    log.Fatalf("Failed to open log file: %v", err)
  }
  defer lf.Close()

  defer log.Init("LoggerExample", *verbose, true, lf).Close()

  log.Info("I'm about to do something!")
  if err := doSomething(); err != nil {
    log.Errorf("Error running doSomething: %v", err)
  }
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

```go
func main() {
    lf, err := os.OpenFile(logPath, …, 0660)
    defer log.Init("foo", *verbose, true, lf).Close()
    log.SetFlags(log.LstdFlags)
}
```

More info: https://golang.org/pkg/log/#pkg-constants