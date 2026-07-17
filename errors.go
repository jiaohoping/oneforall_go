package oneforall

import "errors"

var (
	// ErrPythonNotInstalled is returned when python3 cannot be found on PATH
	// and no explicit path was provided via WithPythonPath.
	ErrPythonNotInstalled = errors.New("python3 is not installed")

	// ErrOneForAllPathNotSet is returned when the path to oneforall.py was not
	// provided via WithOneForAllPath.
	ErrOneForAllPathNotSet = errors.New("oneForAll script path is not set")

	// ErrParseOutput is returned when the result database cannot be found or
	// its contents cannot be parsed.
	ErrParseOutput = errors.New("failed to parse output")
)
