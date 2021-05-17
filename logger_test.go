package log

import (
	"bufio"
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func fakeExit(int) {
	panic("os.Exit called")
}

func TestLoggingBeforeInit(t *testing.T) {
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stderr = w

	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()

	// Reset
	initialize()
	SetLevel(LevelDebug)

	debug := "debug log"
	info := "info log"
	warning := "warning log"
	errL := "error log"
	fatal := "fatal log"

	Debug(debug)
	Debugf(debug)
	Info(info)
	Infof(info)
	Warning(warning)
	Warningf(warning)
	Error(errL)
	Errorf(errL)
	// We don't want os.Exit in a test
	defaultLogger.output(LevelFatal, 0, fatal)
	assert.Panics(t, func() { Fatal(fatal) }, "os exit not called")
	assert.Panics(t, func() { Fatalf(fatal) }, "os exit not called")

	w.Close()
	Close()
	os.Stderr = old

	var b bytes.Buffer
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		b.Write(scanner.Bytes())
	}

	out := b.String()

	for _, txt := range []string{debug, debug, info, info, warning, warning, errL, errL, fatal, fatal} {
		assert.Contains(t, out, txt)
	}
}

func TestInit(t *testing.T) {
	var buf1 bytes.Buffer
	l1 := New(&buf1)
	if !reflect.DeepEqual(l1, defaultLogger) {
		t.Fatal("defaultLogger does not match logger returned by Init")
	}

	// Subsequent runs of Init shouldn't change defaultLogger.
	var buf2 bytes.Buffer
	l2 := New(&buf2)
	if !reflect.DeepEqual(l1, defaultLogger) {
		t.Error("defaultLogger should not have changed")
	}

	// Check log output.
	l1.Info("logger #1")
	l2.Info("logger #2")
	defaultLogger.Info("logger default")

	tests := []struct {
		out  string
		want int
	}{
		{buf1.String(), 2},
		{buf2.String(), 1},
	}

	for i, tt := range tests {
		got := len(strings.Split(strings.TrimSpace(tt.out), "\n"))
		if got != tt.want {
			t.Errorf("logger %d wrong number of lines, want %d, got %d", i+1, tt.want, got)
		}
	}
}

func TestLogWithContextFields(t *testing.T) {
	patch := monkey.Patch(os.Exit, func(int) {})
	defer patch.Unpatch()

	var buf bytes.Buffer
	l := New(&buf)

	lf := LogFields{
		"bool":   true,
		"int":    7,
		"string": "test",
		"struct": struct{ A string }{"aa"},
	}.Add(LogFields{}).Add(LogFields{"second": 2})

	l.SetFlags(Ldisable)
	l.SetLevel(LevelDebug)
	l.With(lf).Info("check field")
	l.With(lf).Debug("check field")
	l.With(lf).Error("check field")
	l.With(lf).Warning("check field")
	l.With(lf).Infof("check field")
	l.With(lf).Debugf("check field")
	l.With(lf).Errorf("check field")
	l.With(lf).Warningf("check field")
	l.With(lf).Fatal("check field")
	l.With(lf).Fatalf("check field")

	s := buf.String()

	for _, line := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		assert.Contains(t, line, "bool=true int=7 second=2 string=test struct={aa} check field")
	}
}
