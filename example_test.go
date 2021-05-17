package log_test

import (
	"os"

	"github.com/bialas1993/log"
)

func ExampleLogsLevel() {
	os.Stderr = os.Stdout
	l := log.NewColorLogger()
	l.SetFlags(log.Ldisable)
	l.SetLevel(log.LevelDebug)

	l.Error("error")
	l.Errorf("%sf", "error")
	l.Debug("debug")
	l.Debugf("%sf", "debug")
	l.Warning("warning")
	l.Warningf("%sf", "warning")
	l.Info("info")
	l.Infof("%sf", "info")
	// without fatal (system exit) and error (another writer os error)

	// Output:
	// ERROR: error
	// ERROR: errorf
	// DEBUG: debug
	// DEBUG: debugf
	// WARN : warning
	// WARN : warningf
	// INFO : info
	// INFO : infof
}
