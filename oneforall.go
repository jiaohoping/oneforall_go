// Package oneforall provides a Go wrapper for calling the OneForAll
// subdomain collection tool and converting its results into structured objects.
package oneforall

import (
	"bufio"
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

// ProgressEventType indicates the kind of event sent on a progress channel.
type ProgressEventType int

const (
	// EventStarted is sent once the OneForAll process has been launched.
	EventStarted ProgressEventType = iota
	// EventStdoutLine is sent for each line written to OneForAll's stdout.
	EventStdoutLine
	// EventCompleted is sent when the scan finishes (successfully or with an error).
	EventCompleted
)

// ProgressEvent carries one update from a RunAsyncWithProgress call.
type ProgressEvent struct {
	Type ProgressEventType
	// Line is populated for EventStdoutLine events.
	Line string
	// Result is populated for EventCompleted events on success.
	Result *Result
	// Err is populated for EventCompleted events on failure.
	Err error
}

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
	// targetFilePath is set by WithTargetFile; the file is read lazily at Run time.
	targetFilePath string
	// runArgs are all --flag [value] pairs placed after "run".
	runArgs []string
	// cleanupFiles lists temporary files to be removed after Run completes.
	cleanupFiles []string
	// initErr holds the first error that occurred during option application.
	// Run() and Validate() check this before proceeding.
	initErr error

	pythonPath    string
	oneforallPath string
	ctx           context.Context

	subdomainFilter func(Subdomain) bool

	streamer     io.Writer
	outputPath   string // merged from WithOutputPath / ToFile
	resultDBPath string // explicit override for the result SQLite path
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

// Validate checks the scanner configuration before running. It verifies that
// no option-application errors occurred, that the python executable and
// oneforall.py script both exist on disk, and that at least one target is set.
func (s *Scanner) Validate() error {
	if s.initErr != nil {
		return s.initErr
	}
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

// Reset clears all target configuration and per-scan state, retaining the
// python/oneforall paths and process-level options (SysProcAttr, Streamer).
// Useful for scanning multiple targets sequentially with the same base config.
func (s *Scanner) Reset() *Scanner {
	s.cleanup()
	s.targetArgs = nil
	s.targets = nil
	s.targetFilePath = ""
	s.runArgs = nil
	s.initErr = nil
	s.outputPath = ""
	s.resultDBPath = ""
	s.subdomainFilter = nil
	return s
}

// Clone returns a new Scanner that shares the same python/oneforall paths and
// process-level options as s. All slices are deep-copied so the clone can be
// independently configured via AddOptions without affecting s.
func (s *Scanner) Clone() *Scanner {
	c := &Scanner{
		modifySysProcAttr: s.modifySysProcAttr,
		pythonPath:        s.pythonPath,
		oneforallPath:     s.oneforallPath,
		ctx:               s.ctx,
		subdomainFilter:   s.subdomainFilter,
		streamer:          s.streamer,
		outputPath:        s.outputPath,
		resultDBPath:      s.resultDBPath,
		outputFormat:      s.outputFormat,
		targetFilePath:    s.targetFilePath,
		initErr:           s.initErr,
	}
	if len(s.targetArgs) > 0 {
		c.targetArgs = append([]string(nil), s.targetArgs...)
	}
	if len(s.targets) > 0 {
		c.targets = append([]string(nil), s.targets...)
	}
	if len(s.runArgs) > 0 {
		c.runArgs = append([]string(nil), s.runArgs...)
	}
	return c
}

// Run executes the OneForAll scan synchronously and returns the parsed result.
// The context passed to NewScanner can be used to cancel or time out the scan.
func (s *Scanner) Run() (*Result, error) {
	defer s.cleanup()

	if s.initErr != nil {
		return nil, s.initErr
	}
	if len(s.targetArgs) == 0 {
		return nil, fmt.Errorf("missing required parameter: --target or --targets")
	}
	if err := s.resolveTargetsFromFile(); err != nil {
		return nil, err
	}

	args := s.Args()
	cmdLine := fmt.Sprintf("%s %s", s.pythonPath, strings.Join(args, " "))
	log.Info().Str("command", cmdLine).Msg("executing command")

	cmd := exec.CommandContext(s.ctx, s.pythonPath, args...)
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

// RunAsyncWithProgress executes the scan in a background goroutine and returns
// a channel that receives ProgressEvents. The channel is closed once the scan
// completes or fails. Events are delivered in order:
//
//	EventStarted → (zero or more EventStdoutLine) → EventCompleted
//
// If a Streamer is also set, each stdout line is forwarded to it in addition
// to being sent as an EventStdoutLine event.
func (s *Scanner) RunAsyncWithProgress() <-chan ProgressEvent {
	ch := make(chan ProgressEvent, 64)

	go func() {
		defer close(ch)
		defer s.cleanup()

		if s.initErr != nil {
			ch <- ProgressEvent{Type: EventCompleted, Err: s.initErr}
			return
		}
		if len(s.targetArgs) == 0 {
			ch <- ProgressEvent{Type: EventCompleted, Err: fmt.Errorf("missing required parameter: --target or --targets")}
			return
		}
		if err := s.resolveTargetsFromFile(); err != nil {
			ch <- ProgressEvent{Type: EventCompleted, Err: err}
			return
		}

		args := s.Args()
		cmdLine := fmt.Sprintf("%s %s", s.pythonPath, strings.Join(args, " "))
		log.Info().Str("command", cmdLine).Msg("executing command (async with progress)")

		cmd := exec.CommandContext(s.ctx, s.pythonPath, args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		if s.modifySysProcAttr != nil {
			s.modifySysProcAttr(cmd.SysProcAttr)
		}

		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			ch <- ProgressEvent{Type: EventCompleted, Err: fmt.Errorf("failed to create stdout pipe: %w", err)}
			return
		}

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err = cmd.Start(); err != nil {
			ch <- ProgressEvent{Type: EventCompleted, Err: fmt.Errorf("failed to start command: %w", err)}
			return
		}
		ch <- ProgressEvent{Type: EventStarted}

		// Read stdout line by line; forward each line to Streamer if set.
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			sc := bufio.NewScanner(stdoutPipe)
			for sc.Scan() {
				line := sc.Text()
				ch <- ProgressEvent{Type: EventStdoutLine, Line: line}
				if s.streamer != nil {
					fmt.Fprintln(s.streamer, line) //nolint:errcheck
				}
			}
		}()

		done := make(chan error, 1)
		go func() {
			wg.Wait()
			done <- cmd.Wait()
		}()

		result := &Result{}
		if err = s.processResult(result, done, &stderr); err != nil {
			log.Error().Err(err).Msg("async command execution failed")
			ch <- ProgressEvent{Type: EventCompleted, Err: err}
			return
		}
		ch <- ProgressEvent{Type: EventCompleted, Result: result}
	}()

	return ch
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

	dbPath := s.resolveResultDBPath()
	log.Debug().
		Strs("targets", s.targets).
		Str("resultDBPath", dbPath).
		Msg("processing result")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Error().Str("resultDBPath", dbPath).Msg("result database file not found")
		return fmt.Errorf("result database file not found: %w", ErrParseOutput)
	}

	if err := result.FromDBMulti(dbPath, s.targets); err != nil {
		log.Error().Err(err).Msg("failed to read result database file")
		return err
	}

	if s.subdomainFilter != nil {
		result.applyFilter(s.subdomainFilter)
	}

	return nil
}

// resolveResultDBPath returns the path to the OneForAll result SQLite database
// using a three-level fallback:
//  1. Explicit override via WithResultDBPath
//  2. Custom output directory set via WithOutputPath / ToFile
//  3. Default: {oneforall.py dir}/results/result.sqlite3
func (s *Scanner) resolveResultDBPath() string {
	if s.resultDBPath != "" {
		return s.resultDBPath
	}
	if s.outputPath != "" {
		return filepath.Join(s.outputPath, RESULT_DB_NAME)
	}
	return filepath.Join(filepath.Dir(s.oneforallPath), RESULT_DB_PATH, RESULT_DB_NAME)
}

// resolveTargetsFromFile reads the target file set by WithTargetFile and
// populates s.targets. It is called lazily at the start of Run / RunAsync.
func (s *Scanner) resolveTargetsFromFile() error {
	if s.targetFilePath == "" {
		return nil
	}
	data, err := os.ReadFile(s.targetFilePath)
	if err != nil {
		return fmt.Errorf("WithTargetFile: cannot read target file %q: %w", s.targetFilePath, err)
	}
	s.targets = nil
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			s.targets = append(s.targets, line)
		}
	}
	if len(s.targets) == 0 {
		return fmt.Errorf("WithTargetFile: target file %q contains no domains", s.targetFilePath)
	}
	return nil
}

func (s *Scanner) cleanup() {
	for _, f := range s.cleanupFiles {
		os.Remove(f) //nolint:errcheck
	}
	s.cleanupFiles = nil
}

// WithResultDBPath explicitly sets the path to the OneForAll SQLite result
// database, overriding the default path resolution. Use this when you have
// configured a custom --path output directory so that results are read from
// the correct location.
func WithResultDBPath(path string) Option {
	return func(s *Scanner) {
		s.resultDBPath = path
	}
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
