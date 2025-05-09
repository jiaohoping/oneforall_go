package oneforall

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05",
	})

	log.Logger = log.With().Caller().Logger()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	zerolog.TimeFieldFormat = time.RFC3339
}

func GetLogger() zerolog.Logger {
	return log.Logger
}

func SetLogLevel(level zerolog.Level) {
	zerolog.SetGlobalLevel(level)
}
