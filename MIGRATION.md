# Migration Guide

## v0.4.x / v1.0.0

### [Breaking] `WithAlive()` removed

The deprecated alias introduced in v0.2.0 has been deleted.

```go
// old (no longer compiles)
scanner.AddOptions(oneforall.WithAlive())

// new
scanner.AddOptions(oneforall.WithValid(true))
```

### [Breaking] `ScanRunnerV1` interface removed

```go
// old (no longer compiles)
var _ oneforall.ScanRunnerV1 = scanner

// new
var _ oneforall.ScanRunner = scanner
```

### [Deprecated] `ToFile` and `Streamer` chainable methods

These still work but will be removed in v2.0.0.

```go
// deprecated
scanner.ToFile("/path/out").Streamer(os.Stdout)

// preferred
scanner.AddOptions(
    oneforall.WithOutputPath("/path/out"),
    oneforall.WithStreamer(os.Stdout),
)
```

### [Non-breaking] New API additions in v0.4.0 / v1.0.0

| Symbol | Description |
|---|---|
| `WithLogger(zerolog.Logger)` | Per-Scanner logger |
| `WithEnv(key, value)` | Set process env var |
| `WithWorkDir(dir)` | Set process working dir |
| `WithStreamer(w)` | Option-based streamer |
| `(*Scanner).RunWithProgress(func(ProgressEvent))` | Sync progress callback |
| `Result.Meta ScanMeta` | Scan timing/command metadata |
| `ScanMeta` | New struct |
| `Result.Diff(previous)` | Incremental diff |
| `SubdomainChange` / `ResultDiff` | Diff result types |

---

## v0.2.x → v0.3.0

No Breaking Changes. All additions are backward-compatible.

### [Non-breaking] `WithTargetFile` now reads file lazily

Previously `WithTargetFile` called `os.ReadFile` at option-apply time. It now
defers the read to `Run()`. If the file does not exist when `Run()` is called,
an error is returned.

### [Non-breaking] `WithTargets` error now propagated

Previously, if `WithTargets` failed to create a temporary file it silently fell
back to the first domain. The error is now stored internally and returned when
`Run()` or `Validate()` is called. No API change is required.

### [Non-breaking] New API additions in v0.3.0

| Symbol | Description |
|---|---|
| `WithResultDBPath(path)` | Override result SQLite path |
| `Result.Unique()` | Remove duplicate subdomain names |
| `Subdomain.CNAMEs() []string` | Parse comma-separated CNAME field |
| `Result.GroupBySource()` | Group subdomains by data source |
| `ResultStats.BySource` | Per-source counts in Stats() |
| `(*Scanner).Reset()` | Clear scan state for reuse |
| `(*Scanner).Clone()` | Deep-copy scanner with shared base config |
| `ProgressEventType` / `ProgressEvent` | Structured progress event types |
| `(*Scanner).RunAsyncWithProgress()` | Channel-based async scan |

---

## v0.1.x → v0.2.0

This release fixes a number of bugs and cleans up the API.  
Several **Breaking Changes** are listed below. Where possible a deprecated
backward-compatible shim has been kept in `deprecated.go` so existing code
continues to compile; the shims will be removed in a future major version.

---

### [Breaking] `WithAlive()` → `WithValid(true)`

**Reason**: `WithAlive()` passed `--alive` to OneForAll, which is not a valid
flag. The correct flag is `--valid True/False`.

**Migration**:

```go
// old
scanner.AddOptions(oneforall.WithAlive())

// new
scanner.AddOptions(oneforall.WithValid(true))
```

The old `WithAlive()` symbol is kept as a deprecated alias that delegates to
`WithValid(true)`. It will be removed in v1.0.0.

---

### [Breaking] `ScanRunner` interface signature change

**Reason**: The original `ScanRunner` interface declared
`Run() (*Result, []string, error)` but `Scanner.Run()` always returned only
`(*Result, error)`. The `warnings []string` was never populated.

**Migration**:

```go
// old interface (kept as ScanRunnerV1 for backward compat)
type ScanRunnerV1 interface {
    Run() (*Result, []string, error)
}

// new interface
type ScanRunner interface {
    Run() (*Result, error)
}
```

If you type-assert against `ScanRunner`, no change is required. If you have
code that implements the old interface with the three-return signature, update
it to `(*Result, error)`.

The old interface is available as `ScanRunnerV1` in `deprecated.go`.

---

### [Breaking] `logger.go` no longer modifies global zerolog state

**Reason**: Libraries must not call `init()` to override the caller's zerolog
configuration. The `init()` function that set a `ConsoleWriter` output and
`InfoLevel` global log level has been removed.

**Impact**: If your application relied on the library to initialise zerolog, you
will need to do so yourself.

**Migration**:

```go
// Add to your application's main() or init()
import (
    "os"
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
}
```

Alternatively use the library helper:

```go
oneforall.SetLogLevel(zerolog.InfoLevel)
```

---

### [Non-breaking] `WithTargets` multi-domain behaviour fixed

**Reason**: The old implementation passed `-t domain1,domain2` (single dash,
comma-separated), which the argument parser silently dropped. Multiple domains
are now written to a temporary file and passed as `--targets <file>`.

No API change is required. The fix is transparent if you already used
`WithTargets("a.com", "b.com")`.

---

### [Non-breaking] `WithOutputPath` and `ToFile` no longer duplicate `--path`

**Reason**: Previously calling both `WithOutputPath` and `ToFile` inserted
`--path` twice into the command line. Now both set the same internal field;
the last call wins and `--path` appears exactly once.

---

### [Non-breaking] New API additions

The following new symbols are available in v0.2.0:

| Symbol | Description |
|---|---|
| `WithValid(bool)` | Replaces `WithAlive()`; maps to `--valid True/False` |
| `WithShow(bool)` | Maps to `--show True/False` |
| `(*Scanner).Validate() error` | Pre-flight check before `Run()` |
| `(*Scanner).RunAsync(func(*Result, error))` | Non-blocking scan |
| `(*Result).FromDBMulti(dbPath string, targets []string) error` | Multi-target DB read |
| `Result.Filter(func(Subdomain) bool) Result` | Functional filter |
| `Result.Alive() Result` | Shortcut for alive-only filter |
| `Result.GroupByModule() map[string][]Subdomain` | Group by discovery module |
| `Result.Stats() ResultStats` | Aggregate statistics |
| `Subdomain.IPs() []string` | Parse comma-separated IP field |
| `Subdomain.IsAlive() bool` | Typed accessor |
| `Subdomain.IsResolved() bool` | Typed accessor |
| `Subdomain.IsRequested() bool` | Typed accessor |
| `Subdomain.IsCDN() bool` | Typed accessor |
| `Subdomain.IsNew() bool` | Typed accessor |
