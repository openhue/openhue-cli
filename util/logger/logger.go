package logger

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func Init(path string) {

	if isProd() {
		// If the file doesn't exist, create it or append to the file
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}

		log.SetOutput(file)
	} else {
		log.Info("Running in dev mode, logs will be output in the console")
	}
}

func isProd() bool {
	return os.Getenv("OPENHUE_ENV") != "dev"
}
