package oneforall

// WithShow controls whether OneForAll prints its results table to stdout
// after the scan completes (--show). Default False.
func WithShow(enable bool) Option {
	val := "False"
	if enable {
		val = "True"
	}
	return func(s *Scanner) {
		s.runArgs = append(s.runArgs, "--show", val)
	}
}
