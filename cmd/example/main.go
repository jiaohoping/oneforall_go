package main

import (
	"context"
	"os"
	"time"

	oneforall "github.com/jiaohoping/oneforall_go"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	scanner, err := oneforall.NewScanner(
		ctx,
		oneforall.WithOneForAllPath("/home/jiao/code/source_code/OneForAll/oneforall.py"),
		oneforall.WithTarget("example.com"),
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create OneForAll scanner")
		return
	}

	scanner.AddOptions(
		oneforall.WithBruteForce(true),
		oneforall.WithPort("small"),
		oneforall.WithDNS(true),
		oneforall.WithRequest(true),
		oneforall.WithTakeover(false),
		oneforall.WithOutputFormat(oneforall.FormatJSON),
	)

	scanner = scanner.Streamer(os.Stdout)

	result, err := scanner.Run()
	if err != nil {
		log.Error().Err(err).Msg("OneForAll scan failed")
		return
	}

	log.Info().Int("count", len(result.Subdomains)).Msg("Scan completed! Found subdomains")

	for i, subdomain := range result.Subdomains {
		if i >= 10 {
			log.Info().Int("remaining", len(result.Subdomains)-10).Msg("More subdomains remaining")
			break
		}
		log.Info().
			Str("subdomain", subdomain.Subdomain).
			Str("ip", subdomain.IP).
			Int("status", subdomain.Status).
			Msg("Found subdomain")
	}
}
