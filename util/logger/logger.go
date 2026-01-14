package logger

import (
	log "github.com/sirupsen/logrus"
	"os"
)

// Init initializes the logger for the application.
// In production mode, logs are written to a file at the given path.
// The file handle is intentionally not closed as it needs to remain open
// for the entire lifetime of the CLI process. The OS will close it on exit.
func Init(path string, level string) {
	if isProd() {
		// If the file doesn't exist, create it or append to the file
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}
		// Note: file is intentionally not closed - see function documentation

		logLevel := log.InfoLevel
		if level != "" {
			parsedLevel, err := log.ParseLevel(level)
			if err != nil {
				log.Warnf("Invalid log level '%s', defaulting to 'info'", level)
			} else {
				logLevel = parsedLevel
			}
		}
		log.SetLevel(logLevel)
		log.SetOutput(file)
	} else {
		log.Info("Running in dev mode, logs will be output in the console")
	}
}

func isProd() bool {
	return os.Getenv("OPENHUE_ENV") != "dev"
}
