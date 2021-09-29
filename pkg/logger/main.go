package logger

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func Setup() {
	if os.Getenv("SIMPLE_TFSWITCH_DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}

	log.SetReportCaller(true)

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		DisableLevelTruncation: true,
	})
}
