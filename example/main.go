package main

import (
	"github.com/bialas1993/log"
	// logger "log"
)

func main() {
	logger := log.New()

	println(
		log.Ldate,
		log.Ltime,
		log.Lmicroseconds,
		log.Lshortfile,
		log.LUTC,
		log.Ldate,
		log.Llongfile,
	)

	// println(
	// 	logger.Ldate,
	// 	logger.Ltime,
	// 	logger.Lmicroseconds,
	// 	logger.Lshortfile,
	// 	logger.LUTC,
	// 	logger.Ldate,
	// 	logger.Llongfile,
	// )

	logger.SetFlags(log.LstdFlags)
	logger.SetLevel(log.LevelError)

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
