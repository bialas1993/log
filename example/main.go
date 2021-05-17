package main

import (
	"github.com/bialas1993/log"
)

func main() {
	logger := log.NewColorLogger()
	logger.SetFlags(log.LstdFlags | log.Lmicroseconds)

	logger.Debug("debug")
	logger.With(log.LogFields{
		"asd":   "bsd",
		"lorem": "ipsum",
		"bang":  10,
		"struct": struct {
			A string
		}{"aaaaaa"},
	}).Info("info")
	logger.Warning("warn")
	logger.Error("error")
	logger.Fatal("fatal")
}
