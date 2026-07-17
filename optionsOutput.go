package oneforall

// OutputFormat represents a supported result export format.
type OutputFormat string

const (
	// FormatCSV exports results as CSV (default).
	FormatCSV OutputFormat = "csv"
	// FormatJSON exports results as JSON.
	FormatJSON OutputFormat = "json"
)

// WithOutputFormat sets the export format for OneForAll results (--fmt).
// Defaults to FormatCSV when not set.
func WithOutputFormat(format OutputFormat) Option {
	return func(s *Scanner) {
		s.outputFormat = string(format)
	}
}

// WithOutputPath sets the output file path for OneForAll results (--path).
// If both WithOutputPath and ToFile are called, the last one applied wins.
func WithOutputPath(path string) Option {
	return func(s *Scanner) {
		s.outputPath = path
	}
}
