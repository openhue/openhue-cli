package logger

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func Init(path string, level string) {
	if isProd() {
		// If the file doesn't exist, create it or append to the file
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}

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
