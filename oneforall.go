// Package oneforall provides a Go wrapper for calling the OneForAll
// subdomain collection tool and converting its results into structured objects.
package oneforall

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
)

// ScanRunner is the interface implemented by Scanner.
type ScanRunner interface {
	Run() (*Result, error)
}

// Scanner holds the configuration for a single OneForAll scan.
type Scanner struct {
	modifySysProcAttr func(*syscall.SysProcAttr)

	// targetArgs are the CLI arguments placed before "run":
	//   ["--target", "foo.com"]  or  ["--targets", "/path/to/file"]
	targetArgs []string
	// targets holds the domain names used for SQLite table lookup after the scan.
	targets []string
	// runArgs are all --flag [value] pairs placed after "run".
	runArgs []string
	// cleanupFiles lists temporary files to be removed after Run completes.
	cleanupFiles []string

	pythonPath    string
	oneforallPath string
	ctx           context.Context

	subdomainFilter func(Subdomain) bool

	streamer     io.Writer
	outputPath   string // merged from WithOutputPath / ToFile
	outputFormat string // defaults to "csv"
}

// Option is a functional option applied to a Scanner.
type Option func(*Scanner)

// NewScanner creates a new Scanner with the supplied options applied.
//
// If WithPythonPath is not provided NewScanner auto-detects python3 on PATH;
// it returns ErrPythonNotInstalled when python3 cannot be found.
// It returns ErrOneForAllPathNotSet when WithOneForAllPath is not provided.
func NewScanner(ctx context.Context, options ...Option) (*Scanner, error) {
	scanner := &Scanner{
		ctx:          ctx,
		outputFormat: "csv",
	}

	for _, option := range options {
		option(scanner)
	}

	if scanner.pythonPath == "" {
		var err error
		scanner.pythonPath, err = exec.LookPath("python3")
		if err != nil {
			return nil, ErrPythonNotInstalled
		}
	}

	if scanner.oneforallPath == "" {
		return nil, ErrOneForAllPathNotSet
	}

	return scanner, nil
}

// Validate checks the scanner configuration before running.
// It verifies that the python executable and oneforall.py script both exist
// on disk, and that at least one target has been configured.
// Call this before Run to surface configuration errors early.
func (s *Scanner) Validate() error {
	if s.pythonPath == "" {
		return ErrPythonNotInstalled
	}
	if _, err := os.Stat(s.pythonPath); err != nil {
		return fmt.Errorf("%w: %w", ErrPythonNotInstalled, err)
	}
	if s.oneforallPath == "" {
		return ErrOneForAllPathNotSet
	}
	if _, err := os.Stat(s.oneforallPath); err != nil {
		return fmt.Errorf("%w: %w", ErrOneForAllPathNotSet, err)
	}
	if len(s.targetArgs) == 0 {
		return fmt.Errorf("missing required parameter: --target or --targets")
	}
	return nil
}

// ToFile sets the path where OneForAll should write its result file (--path).
// If WithOutputPath has also been set, the last call wins.
func (s *Scanner) ToFile(file string) *Scanner {
	s.outputPath = file
	return s
}

// Streamer sets a writer that receives OneForAll's stdout in real time.
func (s *Scanner) Streamer(stream io.Writer) *Scanner {
	s.streamer = stream
	return s
}

// AddOptions applies additional options to the scanner after construction.
func (s *Scanner) AddOptions(options ...Option) *Scanner {
	for _, option := range options {
		option(s)
	}
	return s
}

// Args returns the full CLI argument list (excluding the python executable)
// that will be passed to python3 when Run is called.
func (s *Scanner) Args() []string {
	args := []string{s.oneforallPath}
	args = append(args, s.targetArgs...)
	args = append(args, "run")
	args = append(args, s.runArgs...)
	if s.outputPath != "" {
		args = append(args, "--path", s.outputPath)
	}
	args = append(args, "--fmt", s.outputFormat)
	return args
}

// Run executes the OneForAll scan synchronously and returns the parsed result.
// The context passed to NewScanner can be used to cancel or time out the scan.
func (s *Scanner) Run() (*Result, error) {
	defer s.cleanup()

	if len(s.targetArgs) == 0 {
		return nil, fmt.Errorf("missing required parameter: --target or --targets")
	}

	args := s.Args()
	cmdLine := fmt.Sprintf("%s %s", s.pythonPath, strings.Join(args, " "))
	log.Info().Str("command", cmdLine).Msg("executing command")

	cmd := exec.CommandContext(s.ctx, s.pythonPath, args...)

	// Ensure SysProcAttr is non-nil before passing to user callback.
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	if s.modifySysProcAttr != nil {
		s.modifySysProcAttr(cmd.SysProcAttr)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if s.streamer != nil {
			io.Copy(s.streamer, stdoutPipe) //nolint:errcheck
		} else {
			io.Copy(io.Discard, stdoutPipe) //nolint:errcheck
		}
	}()

	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		wg.Wait()
		done <- cmd.Wait()
	}()

	result := &Result{}
	if err = s.processResult(result, done, &stderr); err != nil {
		log.Error().Err(err).Msg("command execution failed")
		return nil, err
	}

	return result, nil
}

// RunAsync executes the scan in a background goroutine and invokes callback
// when the scan completes. The result pointer passed to callback is nil on error.
func (s *Scanner) RunAsync(callback func(*Result, error)) {
	go func() {
		result, err := s.Run()
		callback(result, err)
	}()
}

func (s *Scanner) processResult(result *Result, done <-chan error, stderr *bytes.Buffer) error {
	errStatus := <-done

	if errStatus != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			const maxStderr = 1024
			if len(stderrStr) > maxStderr {
				stderrStr = stderrStr[:maxStderr] + "... (truncated)"
			}
			return fmt.Errorf("command failed: %w; stderr: %s", errStatus, stderrStr)
		}
		return fmt.Errorf("command failed: %w", errStatus)
	}

	oneforallFolder := filepath.Dir(s.oneforallPath)
	resultDBPath := filepath.Join(oneforallFolder, RESULT_DB_PATH, RESULT_DB_NAME)
	log.Debug().
		Strs("targets", s.targets).
		Str("oneforallFolder", oneforallFolder).
		Str("resultDBPath", resultDBPath).
		Msg("processing result")

	if _, err := os.Stat(resultDBPath); os.IsNotExist(err) {
		log.Error().Str("resultDBPath", resultDBPath).Msg("result database file not found")
		return fmt.Errorf("result database file not found: %w", ErrParseOutput)
	}

	if err := result.FromDBMulti(resultDBPath, s.targets); err != nil {
		log.Error().Err(err).Msg("failed to read result database file")
		return err
	}

	if s.subdomainFilter != nil {
		result.applyFilter(s.subdomainFilter)
	}

	return nil
}

func (s *Scanner) cleanup() {
	for _, f := range s.cleanupFiles {
		os.Remove(f) //nolint:errcheck
	}
	s.cleanupFiles = nil
}

// WithCustomArguments appends raw CLI arguments to the post-"run" argument list.
// Use this for OneForAll flags not yet covered by dedicated Option functions.
func WithCustomArguments(args ...string) Option {
	return func(s *Scanner) {
		s.runArgs = append(s.runArgs, args...)
	}
}

// WithPythonPath sets the path to the python3 executable.
// If not set, NewScanner auto-detects python3 on PATH.
func WithPythonPath(pythonPath string) Option {
	return func(s *Scanner) {
		s.pythonPath = pythonPath
	}
}

// WithOneForAllPath sets the filesystem path to the oneforall.py script.
func WithOneForAllPath(oneforallPath string) Option {
	return func(s *Scanner) {
		s.oneforallPath = oneforallPath
	}
}

// WithFilterSubdomain sets a predicate that filters subdomains from the result.
// Only subdomains for which the predicate returns true are retained.
func WithFilterSubdomain(subdomainFilter func(Subdomain) bool) Option {
	return func(s *Scanner) {
		s.subdomainFilter = subdomainFilter
	}
}

// WithCustomSysProcAttr allows customising the os/exec SysProcAttr before the
// OneForAll process is started (e.g. to set process groups on Linux).
func WithCustomSysProcAttr(f func(*syscall.SysProcAttr)) Option {
	return func(s *Scanner) {
		s.modifySysProcAttr = f
	}
}
