package log_test

import (
	"context"
	"os"

	"github.com/bialas1993/log"
)

func ExampleLogsLevel() {
	os.Stderr = os.Stdout
	l := log.New(nil)
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

func ExampleLogsLevelWithContext() {
	os.Stderr = os.Stdout
	l := log.New(nil).WithContextFields(context.Background(), log.LogFields{
		"_context": "set",
	})
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

	// Output:
	// ERROR: _context=set error
	// ERROR: _context=set errorf
	// DEBUG: _context=set debug
	// DEBUG: _context=set debugf
	// WARN : _context=set warning
	// WARN : _context=set warningf
	// INFO : _context=set info
	// INFO : _context=set infof
}
