# oneforall_go

A Go language wrapper for calling `oneforall` and converting results into structured objects.

## What is OneForAll
[OneForAll](https://github.com/shmilylty/OneForAll) is a powerful subdomain integration tool created by [shmilylty](https://github.com/shmilylty)


## Install 
```bash
 go get github.com/jiaohoping/oneforall_go
```

## Simple example
```go
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

```

The program above outputs:
```txt
2025-05-09 22:24:30 INF main.go:44 > Scan completed! Found subdomains count=6
2025-05-09 22:24:30 INF main.go:55 > Found subdomain ip=96.7.128.198,23.215.0.136,23.192.228.84,96.7.128.175,23.192.228.80,23.215.0.138 status=200 subdomain=example.com
2025-05-09 22:24:30 INF main.go:55 > Found subdomain ip=96.7.129.6,96.7.129.42 status=200 subdomain=www.example.com
2025-05-09 22:24:30 INF main.go:55 > Found subdomain ip=96.7.128.198,23.215.0.136,23.192.228.84,96.7.128.175,23.192.228.80,23.215.0.138 status=200 subdomain=example.com
2025-05-09 22:24:30 INF main.go:55 > Found subdomain ip=96.7.129.6,96.7.129.42 status=200 subdomain=www.example.com
2025-05-09 22:24:30 INF main.go:55 > Found subdomain ip=31.13.94.37 status=0 subdomain=www.google.com.example.com
2025-05-09 22:24:30 INF main.go:55 > Found subdomain ip=31.13.94.37 status=0 subdomain=www.google.com.example.com
```


## External resources
- [OnForAll](https://github.com/shmilylty/OneForAll)
- [nmap for go developers](https://github.com/Ullaakut/nmap), inspiration source
    