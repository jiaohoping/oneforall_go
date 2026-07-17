package oneforall

import (
	"fmt"
	"os"
	"strings"
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
			// Fall back to the first domain only when temp file creation fails.
			s.targets = []string{domains[0]}
			s.targetArgs = []string{"--target", domains[0]}
			return
		}

		for _, d := range domains {
			if _, werr := fmt.Fprintln(f, d); werr != nil {
				f.Close()
				os.Remove(f.Name()) //nolint:errcheck
				s.targets = []string{domains[0]}
				s.targetArgs = []string{"--target", domains[0]}
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
// (--targets). The file is read immediately to populate the target list used
// for result DB lookup; if the file cannot be read at option-apply time the
// path itself is stored as a placeholder.
func WithTargetFile(filePath string) Option {
	return func(s *Scanner) {
		s.targetArgs = []string{"--targets", filePath}

		data, err := os.ReadFile(filePath)
		if err != nil {
			// Store the path as a single target placeholder; lookup may fail
			// at result-parse time if the file truly does not exist.
			s.targets = []string{filePath}
			return
		}

		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				s.targets = append(s.targets, line)
			}
		}
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
