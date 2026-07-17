package oneforall

// This file contains backward-compatible aliases and deprecated types that
// are retained to avoid breaking dependents on a go get upgrade.
// All symbols in this file will be removed in a future major release.

// ScanRunnerV1 is the original ScanRunner interface that included a warnings
// return value. It is kept for code that type-asserts against the old
// interface.
//
// Deprecated: Use ScanRunner instead. The warnings slice was never populated.
type ScanRunnerV1 interface {
	Run() (*Result, []string, error)
}

// WithAlive configures the scan to export only live/valid subdomains.
//
// Deprecated: WithAlive used an incorrect --alive flag that OneForAll does not
// support. Use WithValid(true) instead. This alias is retained for backward
// compatibility and delegates to WithValid(true).
func WithAlive() Option {
	return WithValid(true)
}
