package main

import (
	"github.com/bialas1993/log"
)

func main() {
	logger := log.NewSyslogLogger("white")

	println(
		log.LevelDebug,
		log.LevelInfo,
		log.LevelWaring,
		log.LevelError,
		log.LevelFatal,
	)

	logger.SetLevel(log.LevelDefault | log.LevelDebug)

	logger.Debug("debug")
	logger.With(log.LogFields{
		"asd":   "bsd",
		"lorem": "ipsum",
		"bang":  10,
		"struct": struct {
			A string
		}{"aaaaaa"},
	}).Info("asda")
	logger.Warning("warn")
	logger.Error("error")
	logger.Fatal("fatal")
}
