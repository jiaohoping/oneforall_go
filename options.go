package oneforall

import (
	"strings"
)

func WithTarget(domain string) Option {
	return func(s *Scanner) {
		s.target = domain
		s.args = append(s.args, "--target", domain)
	}
}

func WithTargets(domains ...string) Option {
	domainList := strings.Join(domains, ",")

	return func(s *Scanner) {
		s.args = append(s.args, "-t", domainList)
	}
}

func WithTargetFile(filePath string) Option {
	return func(s *Scanner) {
		s.args = append(s.args, "--targets", filePath)
	}
}

// only export alive subdomains (default False)
func WithAlive() Option {
	return func(s *Scanner) {
		s.args = append(s.args, "--alive")
	}
}

// the port range to reuqest (default small port is 80,443)
func WithPort(portRange string) Option {
	return func(s *Scanner) {
		s.args = append(s.args, "--port", portRange)
	}
}

// use dns resolution (default True)
func WithDNS(enable bool) Option {
	dnsStr := "False"
	if enable {
		dnsStr = "True"
	}
	return func(s *Scanner) {
		s.args = append(s.args, "--dns", dnsStr)
	}
}

// http request subdomains (default True)
func WithRequest(enable bool) Option {
	reqStr := "False"
	if enable {
		reqStr = "True"
	}
	return func(s *Scanner) {
		s.args = append(s.args, "--req", reqStr)
	}
}

// scan subdomain takeover
func WithTakeover(enable bool) Option {
	takeoverStr := "False"
	if enable {
		takeoverStr = "True"
	}
	return func(s *Scanner) {
		s.args = append(s.args, "--takeover", takeoverStr)
	}
}

// use brute module
func WithBruteForce(enable bool) Option {
	bruteStr := "False"
	if enable {
		bruteStr = "True"
	}
	return func(s *Scanner) {
		s.args = append(s.args, "--brute", bruteStr)
	}
}
