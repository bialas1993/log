package log

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/eventlog"
)

type writer struct {
	pri severity
	src string
	el  *eventlog.Log
}

// Write sends a log message to the Event Log.
func (w *writer) Write(b []byte) (int, error) {
	switch w.pri {
	case LevelDebug, LevelInfo:
		return len(b), w.el.Info(1, string(b))
	case LevelWarning:
		return len(b), w.el.Warning(3, string(b))
	case LevelError, LevelPanic, LevelFatal:
		return len(b), w.el.Error(2, string(b))
	}
	return 0, fmt.Errorf("unrecognized severity: %v", w.pri)
}

func (w *writer) Close() error {
	return w.el.Close()
}

func newW(pri severity, src string) (*writer, error) {
	// Continue if we receive "registry key already exists" or if we get
	// ERROR_ACCESS_DENIED so that we can log without administrative permissions
	// for pre-existing eventlog sources.
	if err := eventlog.InstallAsEventCreate(src, eventlog.Info|eventlog.Warning|eventlog.Error); err != nil {
		if !strings.Contains(err.Error(), "registry key already exists") && err != windows.ERROR_ACCESS_DENIED {
			return nil, err
		}
	}
	el, err := eventlog.Open(src)
	if err != nil {
		return nil, err
	}
	return &writer{
		pri: pri,
		src: src,
		el:  el,
	}, nil
}

func setup(src string) (*writer, *writer, *writer, *writer, *writer, error) {
	debugL, err := newW(LevelDebug, src)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	infoL, err := newW(LevelInfo, src)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	warningL, err := newW(LevelWarning, src)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	errL, err := newW(LevelError, src)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	panicL, err := newW(LevelPanic, src)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	return debugL, infoL, warningL, errL, panicL, nil
}
