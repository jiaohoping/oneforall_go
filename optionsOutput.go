package oneforall

type OutputFormat string

const (
	FormatCSV  OutputFormat = "csv"
	FormatJSON OutputFormat = "json"
)

func WithOutputFormat(format OutputFormat) Option {
	return func(s *Scanner) {
		s.outputFormat = string(format)
	}
}

func WithOutputPath(path string) Option {
	return func(s *Scanner) {
		s.args = append(s.args, "--path", path)
	}
}
