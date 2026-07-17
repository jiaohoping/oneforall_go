package main

import (
	"context"
	"os"
	"time"

	oneforall "github.com/jiaohoping/oneforall_go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Configure zerolog (previously handled by the library's init(); now the
	// caller's responsibility).
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05",
	})
	log.Logger = log.With().Caller().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	scanner, err := oneforall.NewScanner(
		ctx,
		oneforall.WithOneForAllPath("/home/jiao/code/source_code/OneForAll/oneforall.py"),
		oneforall.WithTarget("example.com"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create OneForAll scanner")
	}

	scanner.AddOptions(
		oneforall.WithBruteForce(true),
		oneforall.WithPort("small"),
		oneforall.WithDNS(true),
		oneforall.WithRequest(true),
		oneforall.WithTakeover(false),
		oneforall.WithOutputFormat(oneforall.FormatJSON),
	)

	// Validate configuration before starting the scan.
	if err := scanner.Validate(); err != nil {
		log.Fatal().Err(err).Msg("scanner configuration is invalid")
	}

	scanner = scanner.Streamer(os.Stdout)

	result, err := scanner.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("OneForAll scan failed")
	}

	stats := result.Stats()
	log.Info().
		Int("total", stats.Total).
		Int("alive", stats.Alive).
		Int("cdn", stats.CDN).
		Msg("scan completed")

	for i, sub := range result.Subdomains {
		if i >= 10 {
			log.Info().Int("remaining", len(result.Subdomains)-10).Msg("more subdomains remaining")
			break
		}
		log.Info().
			Str("subdomain", sub.Subdomain).
			Strs("ips", sub.IPs()).
			Int("status", sub.Status).
			Bool("alive", sub.IsAlive()).
			Bool("cdn", sub.IsCDN()).
			Msg("found subdomain")
	}
}
