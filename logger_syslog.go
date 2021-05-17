// +build linux darwin freebsd

package log

import (
	"log/syslog"
)

func setup(src string) (*syslog.Writer, *syslog.Writer, *syslog.Writer, *syslog.Writer, *syslog.Writer, error) {
	const facility = syslog.LOG_USER
	dl, err := syslog.New(facility|syslog.LOG_DEBUG, src)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	il, err := syslog.New(facility|syslog.LOG_NOTICE, src)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	wl, err := syslog.New(facility|syslog.LOG_WARNING, src)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	el, err := syslog.New(facility|syslog.LOG_ERR, src)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	pl, err := syslog.New(facility|syslog.LOG_CRIT, src)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	return dl, il, wl, el, pl, nil
}
