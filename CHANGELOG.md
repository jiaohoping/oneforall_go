# Changelog

All notable changes to this project are documented in this file.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)  
Versioning: [Semantic Versioning](https://semver.org/)

---

## [v0.2.0] — 2026-07-17

See [MIGRATION.md](MIGRATION.md) for a detailed upgrade guide from v0.1.x.

### Fixed

- **`WithAlive()` used a non-existent `--alive` flag** — OneForAll's CLI does
  not accept `--alive`; the correct parameter is `--valid True/False`.
  `WithAlive()` is now a deprecated alias for `WithValid(true)`.
- **`WithTargets(...)` silently dropped domains** — The old implementation
  emitted `-t` (single dash) which the argument assembler ignored, causing
  every multi-domain scan to fail with "missing required parameter".
  Fixed: single domain uses `--target`; multiple domains are written to a
  temporary file and passed with `--targets`.
- **`WithTargets` / `WithTargetFile` did not set the target list** — The
  internal `s.target` field was never populated, causing `processResult` to
  query an empty table name from the result database. Both options now
  correctly populate `s.targets`.
- **`ScanRunner` interface mismatch** — The declared interface returned
  `(*Result, []string, error)` but `Scanner.Run()` returned `(*Result, error)`;
  `Scanner` never implemented its own interface. The interface is now
  `Run() (*Result, error)`.
- **`outputFormat` defaulted to empty string** — When `WithOutputFormat` was
  not called, `--fmt ""` was passed to OneForAll. Default is now `"csv"`.
- **`subdomainFilter` was stored but never applied** — The filter set via
  `WithFilterSubdomain` is now applied after the result database is read.
- **Argument assembler silently dropped single-dash flags** — The two-pass
  arg-assembly logic only recognised `--` prefixed flags. The assembler has
  been replaced with a structured approach (separate `targetArgs` / `runArgs`
  slices) that is deterministic and correct.
- **`cmd.SysProcAttr` nil pointer** — `exec.Command` returns a nil
  `SysProcAttr`; passing it to a user callback caused a nil-pointer panic.
  It is now initialised to `&syscall.SysProcAttr{}` before the callback.
- **`WithOutputPath` + `ToFile` produced duplicate `--path`** — Both paths
  now write to the same internal field; `--path` appears exactly once.
- **`stderr` content was captured but never surfaced** — On process failure
  the stderr output (up to 1 KiB) is now appended to the returned error.
- **`fmt.Errorf` used `%v` instead of `%w`** — All internal error wrapping now
  uses `%w`, enabling `errors.Is` / `errors.As` unwrapping by callers.

### Added

- **`WithValid(bool)`** — Correct replacement for `WithAlive()` (maps to
  `--valid True/False`).
- **`WithShow(bool)`** — Exposes OneForAll's `--show` flag.
- **`(*Scanner).Validate() error`** — Pre-flight check that verifies the
  python executable and oneforall.py script exist on disk and that at least
  one target is configured.
- **`(*Scanner).RunAsync(func(*Result, error))`** — Asynchronous scan;
  replaces the internal dead-code async path.
- **`(*Result).FromDB(dbPath, target string) error`** — Now prefers the
  `<target>_now_result` table (standard OneForAll output) and falls back to
  the plain `<target>` table.
- **`(*Result).FromDBMulti(dbPath string, targets []string) error`** — Reads
  results for multiple target domains from a single SQLite database.
- **`Result.Filter(func(Subdomain) bool) Result`** — Functional filter over
  the subdomain list.
- **`Result.Alive() Result`** — Shortcut that keeps only alive subdomains.
- **`Result.GroupByModule() map[string][]Subdomain`** — Groups subdomains by
  the discovery module that found them.
- **`Result.Stats() ResultStats`** — Aggregate statistics: total, alive, CDN,
  resolved, new, per-module counts.
- **`Subdomain.IPs() []string`** — Parses the comma-separated `IP` field into
  a typed slice.
- **`Subdomain.IsAlive() bool`**, **`IsResolved() bool`**,
  **`IsRequested() bool`**, **`IsCDN() bool`**, **`IsNew() bool`** — Typed
  boolean accessors for the 0/1 integer fields.
- **43 unit tests** covering scanner options, argument assembly, result
  filtering/aggregation, and SQLite parsing.

### Changed

- **`logger.go` no longer calls `init()`** — Libraries must not modify global
  zerolog state. Callers are responsible for configuring zerolog in their own
  `main` or `init`. See [MIGRATION.md](MIGRATION.md) for examples.
- **`errors.go`** — Removed unused `ErrModuleNotFound` and
  `ErrPermissionDenied`; `ErrParseOutput` is now returned when the result
  database is missing.

### Deprecated

- **`WithAlive()`** — Use `WithValid(true)`. Will be removed in v1.0.0.
- **`ScanRunnerV1`** — The old interface with a `warnings []string` return.
  Use `ScanRunner`. Will be removed in v1.0.0.

---

## [v1.0.0] — 2026-07-21

### Breaking Changes

- **`WithAlive()` removed** — Deleted from `deprecated.go`. Use `WithValid(true)`.
- **`ScanRunnerV1` interface removed** — Deleted. Use `ScanRunner`.

### Deprecated (will be removed in v2.0.0)

- **`ToFile(path)`** — Use `WithOutputPath(path)` instead.
- **`Streamer(w)`** — Use `WithStreamer(w)` instead.

### Added (from v0.4.0 features bundled into v1.0.0)

See v0.4.0 entry below for the full feature list.

---

## [v0.4.0] — 2026-07-21

### Fixed

- None (all defects addressed in prior releases).

### Added

- **CI/CD** — `.github/workflows/ci.yml` runs `go build`, `go vet`, `go test -race`,
  and `golangci-lint` on every push and pull request.
- **`.golangci.yml`** — Centralised lint configuration; removes scattered
  `//nolint:errcheck` comments.
- **`Result.Diff(previous Result) ResultDiff`** — Computes the difference between
  two consecutive scan results. Returns `Added`, `Removed`, and `Changed`
  (`SubdomainChange`) slices. Changed detection covers IP, Status, Alive, CDN.
- **`SubdomainChange` / `ResultDiff`** — New types supporting `Result.Diff`.
- **`WithLogger(zerolog.Logger) Option`** — Per-Scanner logger. All internal log
  messages use the scanner's own logger instead of the global one. Defaults to
  `log.Logger` at `NewScanner` time. Enables distinguishing logs from concurrent
  scanners.
- **`WithEnv(key, value string) Option`** — Adds a `KEY=VALUE` entry to the
  OneForAll process environment. Additive; can be called multiple times.
- **`WithWorkDir(dir string) Option`** — Sets the working directory for the
  OneForAll process.
- **`WithStreamer(io.Writer) Option`** — Option-style alternative to the chainable
  `Streamer(w)` method.
- **`(*Scanner).RunWithProgress(onEvent func(ProgressEvent)) (*Result, error)`** —
  Synchronous wrapper around `RunAsyncWithProgress`; calls `onEvent` for each
  event in the caller's goroutine.
- **`Result.Meta ScanMeta`** — Scan metadata populated by `Run`,
  `RunWithProgress`, and `RunAsyncWithProgress`: `Targets`, `StartedAt`,
  `Duration`, `Command`.
- **`ScanMeta`** — New type with `Targets []string`, `StartedAt time.Time`,
  `Duration time.Duration`, `Command string`.
- **Concurrency documentation** — Package-level GoDoc explicitly states that
  Scanner is not safe for concurrent use and points to `Clone` as the remedy.
- **15 new tests** across `scanner_test.go`, `models_test.go`, and `async_test.go`.
  Total: **71 passing tests**.

### Deprecated

- **`ToFile(path)`** (chainable method) — Use `WithOutputPath(path)`.
- **`Streamer(w)`** (chainable method) — Use `WithStreamer(w)`.

---

## [v0.3.0] — 2026-07-17

### Fixed

- **Result DB path not honouring custom `--path`** — `processResult` previously
  always read from `{oneforall.py dir}/results/result.sqlite3` regardless of
  any custom output directory set via `WithOutputPath` or `ToFile`. The path is
  now resolved with a three-level fallback: explicit `WithResultDBPath` override
  → inferred from `outputPath` → default location.
- **Option errors silently ignored** — `WithTargets` used to fall back to the
  first domain when temporary file creation failed, with no indication of the
  failure. Errors are now stored in an internal `initErr` field and returned by
  `Run()` and `Validate()`.
- **`WithTargetFile` read file eagerly at option-apply time** — The target file
  is now read lazily when `Run()` is called, so it does not need to exist when
  the option is applied.

### Added

- **`WithResultDBPath(path string) Option`** — explicit override for the result
  SQLite path; useful when using a custom `--path` output directory.
- **`Result.Unique() Result`** — removes duplicate subdomain names; keeps the
  first occurrence. Useful after multi-target or multi-module scans.
- **`Subdomain.CNAMEs() []string`** — parses the comma-separated `CNAME` field
  into a typed slice, symmetric with the existing `IPs()` method.
- **`Result.GroupBySource() map[string][]Subdomain`** — groups subdomains by
  their `Source` field (the data source that found them).
- **`ResultStats.BySource map[string]int`** — per-source counts in `Stats()`.
- **`(*Scanner).Reset() *Scanner`** — clears target configuration and per-scan
  state while retaining python/oneforall paths and process-level options.
  Enables Scanner reuse for sequential scans without reconstruction.
- **`(*Scanner).Clone() *Scanner`** — returns a new Scanner with deep-copied
  slices sharing the same base configuration. Enables parallel scans with
  per-scan option overrides.
- **`ProgressEventType` / `ProgressEvent`** — structured event types for
  real-time scan progress.
- **`(*Scanner).RunAsyncWithProgress() <-chan ProgressEvent`** — non-blocking
  scan that streams `EventStarted`, `EventStdoutLine`, and `EventCompleted`
  events through a buffered channel. The channel is closed when the scan ends.
  If a `Streamer` is also configured, each stdout line is forwarded to it.
- **Additional tests** — 25 new test cases covering: `Reset`, `Clone`,
  `WithTargetFile` deferred read, `initErr` propagation, `WithResultDBPath`,
  `Unique`, `CNAMEs`, `GroupBySource`, `Stats.BySource`, and the full
  `RunAsyncWithProgress` event lifecycle (in `async_test.go`).

---

## [v0.1.0] — initial release
