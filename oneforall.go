// Package oneforall 提供了对 OneForAll 子域名收集工具的 Go 语言封装
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

type ScanRunner interface {
	Run() (result *Result, warnings []string, err error)
}

type Scanner struct {
	modifySysProcAttr func(*syscall.SysProcAttr)

	args          []string
	target        string
	pythonPath    string
	oneforallPath string
	ctx           context.Context

	subdomainFilter func(Subdomain) bool

	doneAsync    chan error
	streamer     io.Writer
	toFile       *string
	outputFormat string
}

type Option func(*Scanner)

func NewScanner(ctx context.Context, options ...Option) (*Scanner, error) {
	scanner := &Scanner{
		doneAsync: nil,
		streamer:  nil,
		ctx:       ctx,
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

func (s *Scanner) ToFile(file string) *Scanner {
	s.toFile = &file
	return s
}

func (s *Scanner) Streamer(stream io.Writer) *Scanner {
	s.streamer = stream
	return s
}

func (s *Scanner) Run() (result *Result, err error) {
	var stdoutPipe io.ReadCloser
	var stderr bytes.Buffer

	args := []string{s.oneforallPath}

	targetFound := false
	for i := 0; i < len(s.args); i++ {
		if i+1 < len(s.args) && (s.args[i] == "--target" || s.args[i] == "--targets") {
			args = append(args, s.args[i], s.args[i+1])
			targetFound = true
			i++
		}
	}

	if !targetFound {
		return nil, fmt.Errorf("missing required parameter: --target or --targets")
	}

	args = append(args, "run")

	// add all other parameters
	for i := 0; i < len(s.args); i++ {
		if i+1 < len(s.args) && (s.args[i] == "--target" || s.args[i] == "--targets") {
			i++ // skip processed target parameters
			continue
		}

		// add non-target parameters
		if strings.HasPrefix(s.args[i], "--") {
			if i+1 < len(s.args) && !strings.HasPrefix(s.args[i+1], "--") {
				args = append(args, s.args[i], s.args[i+1])
				i++
			} else {
				args = append(args, s.args[i])
			}
		}
	}

	// if output file is set, add corresponding parameters
	if s.toFile != nil {
		args = append(args, "--path", *s.toFile)
	}

	// set output format
	args = append(args, "--fmt", s.outputFormat)

	// print full command line for debugging
	cmdLine := fmt.Sprintf("%s %s", s.pythonPath, strings.Join(args, " "))
	log.Info().Str("command", cmdLine).Msg("executing command")

	// prepare OneForAll process
	cmd := exec.CommandContext(s.ctx, s.pythonPath, args...)
	if s.modifySysProcAttr != nil {
		s.modifySysProcAttr(cmd.SysProcAttr)
	}

	stdoutPipe, err = cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	cmd.Stderr = &stderr

	var wg sync.WaitGroup

	if s.streamer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			io.Copy(s.streamer, stdoutPipe)
		}()
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			io.Copy(io.Discard, stdoutPipe)
		}()
	}

	// run OneForAll process
	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %v", err)
	}

	// wait for all IO operations to complete
	done := make(chan error, 1)
	go func() {
		wg.Wait()
		done <- cmd.Wait()
	}()

	// process result
	result = &Result{}

	if s.doneAsync != nil {
		go func() {
			s.doneAsync <- s.processResult(result, done)
		}()
		return result, nil
	}
	err = s.processResult(result, done)
	if err != nil {
		log.Error().Err(err).Msg("Command execution failed")
		return nil, err
	}

	return result, nil
}

func (s *Scanner) processResult(result *Result, done chan error) error {

	errStatus := <-done

	oneforallFolder := filepath.Dir(s.oneforallPath)
	resultDBPath := filepath.Join(oneforallFolder, RESULT_DB_PATH, RESULT_DB_NAME)
	log.Debug().Str("target", s.target).Str("oneforallPath", s.oneforallPath).Str("oneforallFolder", oneforallFolder).Str("resultDBPath", resultDBPath).Msg("processing result")

	// 1. check error
	if errStatus != nil {
		return errStatus
	}

	//  2. check if sqlite database file exists
	if _, err := os.Stat(resultDBPath); os.IsNotExist(err) {
		log.Error().Str("resultDBPath", resultDBPath).Msg("result database file not found")
		return fmt.Errorf("result database file not found")
	}

	// 3. read sqlite database file
	err := result.FromDB(resultDBPath, s.target)
	if err != nil {
		log.Error().Err(err).Msg("failed to read result database file")
		return err
	}

	return nil
}

// add options
func (s *Scanner) AddOptions(options ...Option) *Scanner {
	for _, option := range options {
		option(s)
	}
	return s
}

// return args list
func (s *Scanner) Args() []string {
	return s.args
}

// set custom arguments
func WithCustomArguments(args ...string) Option {
	return func(s *Scanner) {
		s.args = append(s.args, args...)
	}
}

// set python path
func WithPythonPath(pythonPath string) Option {
	return func(s *Scanner) {
		s.pythonPath = pythonPath
	}
}

// set oneforall path
func WithOneForAllPath(oneforallPath string) Option {
	return func(s *Scanner) {
		s.oneforallPath = oneforallPath
	}
}

// set subdomain filter
func WithFilterSubdomain(subdomainFilter func(Subdomain) bool) Option {
	return func(s *Scanner) {
		s.subdomainFilter = subdomainFilter
	}
}

// set custom sys proc attr
func WithCustomSysProcAttr(f func(*syscall.SysProcAttr)) Option {
	return func(s *Scanner) {
		s.modifySysProcAttr = f
	}
}
