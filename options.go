package oneforall

import (
	"fmt"
	"os"
)

// WithTarget sets a single target domain for the scan (--target).
func WithTarget(domain string) Option {
	return func(s *Scanner) {
		s.targets = []string{domain}
		s.targetArgs = []string{"--target", domain}
	}
}

// WithTargets sets one or more target domains.
//
// For a single domain it maps to --target. For multiple domains a temporary
// file containing one domain per line is created and --targets is used; the
// file is automatically removed when Run completes.
//
// If the temporary file cannot be created the error is stored internally and
// returned when Run or Validate is called.
func WithTargets(domains ...string) Option {
	return func(s *Scanner) {
		if len(domains) == 0 {
			return
		}
		if len(domains) == 1 {
			s.targets = []string{domains[0]}
			s.targetArgs = []string{"--target", domains[0]}
			return
		}

		f, err := os.CreateTemp("", "oneforall-targets-*.txt")
		if err != nil {
			s.initErr = fmt.Errorf("WithTargets: failed to create temp file: %w", err)
			return
		}

		for _, d := range domains {
			if _, werr := fmt.Fprintln(f, d); werr != nil {
				f.Close()
				os.Remove(f.Name()) //nolint:errcheck
				s.initErr = fmt.Errorf("WithTargets: failed to write temp file: %w", werr)
				return
			}
		}
		f.Close()

		s.targets = append([]string(nil), domains...)
		s.targetArgs = []string{"--targets", f.Name()}
		s.cleanupFiles = append(s.cleanupFiles, f.Name())
	}
}

// WithTargetFile sets a file containing one domain per line as scan targets
// (--targets). The file is read lazily when Run is called, so it does not need
// to exist at the time this option is applied.
func WithTargetFile(filePath string) Option {
	return func(s *Scanner) {
		s.targetFilePath = filePath
		s.targetArgs = []string{"--targets", filePath}
	}
}

// WithValid controls whether only valid/live subdomains are exported (--valid).
// Default is False (export all subdomains).
func WithValid(enable bool) Option {
	val := "False"
	if enable {
		val = "True"
	}
	return func(s *Scanner) {
		s.runArgs = append(s.runArgs, "--valid", val)
	}
}

// WithPort sets the port range used when probing subdomains (--port).
// Accepted values: "default", "small", "large".
func WithPort(portRange string) Option {
	return func(s *Scanner) {
		s.runArgs = append(s.runArgs, "--port", portRange)
	}
}

// WithDNS controls whether DNS resolution is performed (--dns). Default True.
func WithDNS(enable bool) Option {
	val := "False"
	if enable {
		val = "True"
	}
	return func(s *Scanner) {
		s.runArgs = append(s.runArgs, "--dns", val)
	}
}

// WithRequest controls whether HTTP requests are sent to discovered subdomains
// (--req). Default True.
func WithRequest(enable bool) Option {
	val := "False"
	if enable {
		val = "True"
	}
	return func(s *Scanner) {
		s.runArgs = append(s.runArgs, "--req", val)
	}
}

// WithTakeover controls whether subdomain takeover checks are performed
// (--takeover). Default False.
func WithTakeover(enable bool) Option {
	val := "False"
	if enable {
		val = "True"
	}
	return func(s *Scanner) {
		s.runArgs = append(s.runArgs, "--takeover", val)
	}
}

// WithBruteForce controls whether the brute-force subdomain module is enabled
// (--brute). Default False.
func WithBruteForce(enable bool) Option {
	val := "False"
	if enable {
		val = "True"
	}
	return func(s *Scanner) {
		s.runArgs = append(s.runArgs, "--brute", val)
	}
}
