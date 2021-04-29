package main

import (
	"github.com/bialas1993/log"
	// logger "log"
)

const logPath = "/tmp/example.log"

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
	logger.SetLevel(5) //TODO
	logger.With(log.LogFields{
		"asd":   "bsd",
		"lorem": "ipsum",
		"bang":  10,
		"struct": struct {
			A string
		}{"aaaaaa"},
	}).Info("asda")
	logger.Error("error")
	logger.Warning("warn")

}
