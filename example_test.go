package log_test

import "github.com/bialas1993/log"

func ExampleNewLogger() {
	logger := log.NewStdLogger()

	logger.SetLevel(log.LevelDefault)
	logger.SetFlags(0)

	logger.Info("Info")
	logger.Debug("Debug")

	// Output:
	// INFO : Info
}
