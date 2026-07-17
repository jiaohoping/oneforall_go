# oneforall_go

A Go library that wraps [OneForAll](https://github.com/shmilylty/OneForAll) and
converts its results into structured Go objects.

## What is OneForAll

[OneForAll](https://github.com/shmilylty/OneForAll) is a powerful subdomain
collection tool created by [shmilylty](https://github.com/shmilylty).

## Install

```bash
go get github.com/jiaohoping/oneforall_go
```

## Quick example

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
        oneforall.WithOneForAllPath("/path/to/OneForAll/oneforall.py"),
        oneforall.WithTarget("example.com"),
    )
    if err != nil {
        log.Fatal().Err(err).Msg("failed to create scanner")
    }

    scanner.AddOptions(
        oneforall.WithBruteForce(true),
        oneforall.WithPort("small"),
        oneforall.WithDNS(true),
        oneforall.WithRequest(true),
        oneforall.WithTakeover(false),
        oneforall.WithOutputFormat(oneforall.FormatJSON),
    )

    // Validate configuration before running (optional but recommended)
    if err := scanner.Validate(); err != nil {
        log.Fatal().Err(err).Msg("invalid scanner configuration")
    }

    result, err := scanner.Streamer(os.Stdout).Run()
    if err != nil {
        log.Fatal().Err(err).Msg("scan failed")
    }

    stats := result.Stats()
    log.Info().
        Int("total", stats.Total).
        Int("alive", stats.Alive).
        Msg("scan completed")

    for i, sub := range result.Subdomains {
        if i >= 10 {
            log.Info().Int("remaining", len(result.Subdomains)-10).Msg("more subdomains")
            break
        }
        log.Info().
            Str("subdomain", sub.Subdomain).
            Strs("ips", sub.IPs()).
            Int("status", sub.Status).
            Bool("alive", sub.IsAlive()).
            Msg("found subdomain")
    }
}
```

## Filtering and aggregation

```go
// Keep only alive subdomains
alive := result.Alive()

// Custom filter
cdnFree := result.Filter(func(s oneforall.Subdomain) bool {
    return !s.IsCDN()
})

// Group by discovery module
byModule := result.GroupByModule()
for module, subs := range byModule {
    log.Info().Str("module", module).Int("count", len(subs)).Msg("")
}

// Aggregate statistics
stats := result.Stats()
// stats.Total, stats.Alive, stats.CDN, stats.Resolved, stats.New, stats.ByModule
```

## Multiple targets

```go
// Inline multiple domains — a temporary file is created automatically
scanner, _ := oneforall.NewScanner(ctx,
    oneforall.WithOneForAllPath("/path/to/oneforall.py"),
    oneforall.WithTargets("a.com", "b.com", "c.com"),
)

// Or point to a file containing one domain per line
scanner, _ = oneforall.NewScanner(ctx,
    oneforall.WithOneForAllPath("/path/to/oneforall.py"),
    oneforall.WithTargetFile("/path/to/domains.txt"),
)
```

## Asynchronous scan

```go
scanner.RunAsync(func(result *oneforall.Result, err error) {
    if err != nil {
        log.Error().Err(err).Msg("scan failed")
        return
    }
    log.Info().Int("count", len(result.Subdomains)).Msg("async scan done")
})
```

## Available options

| Option | OneForAll flag | Default |
|---|---|---|
| `WithTarget(domain)` | `--target` | — |
| `WithTargets(domains...)` | `--target` / `--targets` | — |
| `WithTargetFile(path)` | `--targets` | — |
| `WithBruteForce(bool)` | `--brute` | False |
| `WithDNS(bool)` | `--dns` | True |
| `WithRequest(bool)` | `--req` | True |
| `WithTakeover(bool)` | `--takeover` | False |
| `WithValid(bool)` | `--valid` | False |
| `WithPort(range)` | `--port` | default |
| `WithShow(bool)` | `--show` | False |
| `WithOutputFormat(fmt)` | `--fmt` | csv |
| `WithOutputPath(path)` | `--path` | — |
| `WithCustomArguments(args...)` | any | — |

## Upgrading from v0.1.x

See [MIGRATION.md](MIGRATION.md) for the full upgrade guide.

**Quick summary of breaking changes:**

- `WithAlive()` → `WithValid(true)` (old symbol is kept as deprecated alias)
- `ScanRunner` interface no longer has a `warnings []string` return value
- `logger.go` no longer auto-initialises zerolog; configure it in your `main`

## External resources

- [OneForAll](https://github.com/shmilylty/OneForAll)
- [nmap for Go developers](https://github.com/Ullaakut/nmap) — inspiration source
