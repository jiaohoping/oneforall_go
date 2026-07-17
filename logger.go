package oneforall

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// GetLogger returns the package-level zerolog logger used by this package for
// internal logging.
//
// To customise log format, output destination, or level configure zerolog in
// your application's main function before creating any Scanner. For example:
//
//	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
//	zerolog.SetGlobalLevel(zerolog.DebugLevel)
func GetLogger() zerolog.Logger {
	return log.Logger
}

// SetLogLevel sets the global minimum log level for zerolog.
func SetLogLevel(level zerolog.Level) {
	zerolog.SetGlobalLevel(level)
}
