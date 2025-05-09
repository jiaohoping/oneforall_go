package oneforall

import "errors"

var (
	// ErrPythonNotInstalled indicates Python is not installed
	ErrPythonNotInstalled = errors.New("python3 is not installed")

	// ErrOneForAllPathNotSet indicates OneForAll path is not set
	ErrOneForAllPathNotSet = errors.New("oneForAll script path is not set")

	// ErrModuleNotFound indicates required Python modules are missing
	ErrModuleNotFound = errors.New("required Python modules are missing")

	// ErrPermissionDenied indicates insufficient permissions
	ErrPermissionDenied = errors.New("permission denied")

	// ErrParseOutput indicates output parsing failure
	ErrParseOutput = errors.New("failed to parse output")
)
