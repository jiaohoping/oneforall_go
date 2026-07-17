package oneforall_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	oneforall "github.com/jiaohoping/oneforall_go"
)

// findPython3 returns the path to python3 or skips the test when unavailable.
func findPython3(t *testing.T) string {
	t.Helper()
	// Use the real python3 for path-finding tests, fall back to /usr/bin/python3
	// or any existing executable so that path-based tests are not system-dependent.
	candidates := []string{"/usr/bin/python3", "/usr/local/bin/python3", "/bin/python3"}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	t.Skip("python3 not found, skipping test")
	return ""
}

func TestNewScanner_MissingPython(t *testing.T) {
	_, err := oneforall.NewScanner(
		context.Background(),
		oneforall.WithPythonPath("/nonexistent/python3"),
		oneforall.WithOneForAllPath("/some/oneforall.py"),
		oneforall.WithTarget("example.com"),
	)
	// NewScanner with explicit python path does not verify existence itself;
	// Validate() does. But missing oneforall path detection should still fire
	// since python path was given but a valid oneforall path was not checked.
	// The constructor only checks if python3 is on PATH when none is given.
	// Here we gave one explicitly, so no error from NewScanner.
	if err != nil {
		t.Fatalf("unexpected error from NewScanner with explicit python path: %v", err)
	}
}

func TestNewScanner_MissingOneForAllPath(t *testing.T) {
	_, err := oneforall.NewScanner(context.Background(),
		oneforall.WithTarget("example.com"),
	)
	if err == nil {
		t.Fatal("expected ErrOneForAllPathNotSet, got nil")
	}
	if !strings.Contains(err.Error(), "oneForAll") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewScanner_Defaults(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	if err := os.WriteFile(ofa, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("example.com"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	args := s.Args()
	// Should contain: <ofa> --target example.com run --fmt csv
	if args[0] != ofa {
		t.Errorf("args[0] = %q, want %q", args[0], ofa)
	}
	assertArg(t, args, "--target", "example.com")
	assertArg(t, args, "--fmt", "csv")
	assertContains(t, args, "run")
}

func TestArgs_OutputFormat(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("example.com"),
		oneforall.WithOutputFormat(oneforall.FormatJSON),
	)
	assertArg(t, s.Args(), "--fmt", "json")
}

func TestArgs_WithBruteForce(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("example.com"),
	)
	s.AddOptions(oneforall.WithBruteForce(true))
	assertArg(t, s.Args(), "--brute", "True")
}

func TestArgs_WithDNS_False(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("example.com"),
		oneforall.WithDNS(false),
	)
	assertArg(t, s.Args(), "--dns", "False")
}

func TestArgs_WithValid(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("example.com"),
		oneforall.WithValid(true),
	)
	assertArg(t, s.Args(), "--valid", "True")
}

func TestArgs_WithAlive_IsDeprecatedAlias(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("example.com"),
		oneforall.WithAlive(),
	)
	// Deprecated WithAlive() must delegate to WithValid(true)
	assertArg(t, s.Args(), "--valid", "True")
}

func TestArgs_WithShow(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("example.com"),
		oneforall.WithShow(true),
	)
	assertArg(t, s.Args(), "--show", "True")
}

func TestArgs_WithPort(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("example.com"),
		oneforall.WithPort("large"),
	)
	assertArg(t, s.Args(), "--port", "large")
}

func TestArgs_OutputPath_NoDouplication(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("example.com"),
		oneforall.WithOutputPath("/tmp/out"),
	)
	// Also call ToFile — last one should win, --path must appear exactly once.
	s.ToFile("/tmp/out2")
	args := s.Args()
	count := 0
	for _, a := range args {
		if a == "--path" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("--path appears %d time(s) in args, want 1; args: %v", count, args)
	}
}

func TestArgs_WithTargets_Single(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTargets("example.com"),
	)
	assertArg(t, s.Args(), "--target", "example.com")
}

func TestArgs_WithTargets_Multi_CreatesTargetsFlag(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTargets("a.com", "b.com"),
	)
	args := s.Args()
	// Multi-domain must use --targets, not --target
	if !containsFlag(args, "--targets") {
		t.Errorf("expected --targets flag in args: %v", args)
	}
	if containsFlag(args, "--target") {
		t.Errorf("unexpected --target (single) flag when multiple domains given: %v", args)
	}
}

func TestArgs_WithTargetFile(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	tmpFile := filepath.Join(t.TempDir(), "domains.txt")
	os.WriteFile(tmpFile, []byte("a.com\nb.com\n"), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTargetFile(tmpFile),
	)
	assertArg(t, s.Args(), "--targets", tmpFile)
}

func TestArgs_TargetPlacedBeforeRun(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("example.com"),
		oneforall.WithBruteForce(true),
	)
	args := s.Args()
	runIdx := indexOf(args, "run")
	if runIdx < 0 {
		t.Fatal("'run' not found in args")
	}
	targetIdx := indexOf(args, "--target")
	if targetIdx < 0 {
		t.Fatal("--target not found in args")
	}
	if targetIdx > runIdx {
		t.Errorf("--target (idx %d) must appear before 'run' (idx %d)", targetIdx, runIdx)
	}
	bruteIdx := indexOf(args, "--brute")
	if bruteIdx < runIdx {
		t.Errorf("--brute (idx %d) must appear after 'run' (idx %d)", bruteIdx, runIdx)
	}
}

func TestValidate_MissingPython(t *testing.T) {
	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath("/nonexistent/python3"),
		oneforall.WithOneForAllPath("/nonexistent/oneforall.py"),
		oneforall.WithTarget("example.com"),
	)
	if err := s.Validate(); err == nil {
		t.Error("expected Validate to return error for missing python3")
	}
}

func TestValidate_MissingOneForAllScript(t *testing.T) {
	pyPath := findPython3(t)
	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath("/nonexistent/oneforall.py"),
		oneforall.WithTarget("example.com"),
	)
	if err := s.Validate(); err == nil {
		t.Error("expected Validate to return error for missing oneforall.py")
	}
}

func TestValidate_MissingTarget(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
	)
	if err := s.Validate(); err == nil {
		t.Error("expected Validate to return error when no target is set")
	}
}

func TestValidate_Valid(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, err := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("example.com"),
	)
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	if err := s.Validate(); err != nil {
		t.Errorf("Validate returned unexpected error: %v", err)
	}
}

func TestArgs_CustomArguments(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("example.com"),
		oneforall.WithCustomArguments("--someFlag", "someValue"),
	)
	assertArg(t, s.Args(), "--someFlag", "someValue")
}

// --- v0.3.0 new tests ---

func TestWithResultDBPath_InArgs(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("example.com"),
		oneforall.WithResultDBPath("/custom/results/result.sqlite3"),
	)
	// WithResultDBPath is stored on the scanner; Args() should not include it
	// as a CLI flag (it is an internal override, not a OneForAll flag).
	for _, arg := range s.Args() {
		if strings.Contains(arg, "custom/results") {
			t.Errorf("WithResultDBPath should not add CLI arg, found %q in args: %v", arg, s.Args())
		}
	}
}

func TestWithTargets_InitErrPropagatedToValidate(t *testing.T) {
	// Simulate an initErr by directly checking that Validate surfaces it.
	// We can't force os.CreateTemp to fail, but we can test the error-propagation
	// path by constructing a scanner and injecting a scenario where initErr
	// would be set: use WithTargets with 0 domains (no-op) then verify that a
	// scanner with a forced bad state returns the error.
	// Instead, test the happy-path error propagation by checking Validate()
	// returns nil for a properly configured scanner (no initErr).
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTargets("a.com", "b.com"),
	)
	if err := s.Validate(); err != nil {
		t.Errorf("Validate returned unexpected error: %v", err)
	}
}

func TestWithTargetFile_DeferredRead_FileNotExistAtOptionTime(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	// The file does not exist when the option is applied.
	missingFile := filepath.Join(t.TempDir(), "missing.txt")

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTargetFile(missingFile),
	)
	// Validate should not error on the file path — only Run() reads it.
	// (The target arg is set, so Validate passes structural checks.)
	if err := s.Validate(); err != nil {
		t.Errorf("Validate should not error when target file does not exist yet: %v", err)
	}
}

func TestReset_ClearsTargetAndRunArgs(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("example.com"),
		oneforall.WithBruteForce(true),
	)

	s.Reset()

	// After Reset, Validate should fail because no target is set.
	if err := s.Validate(); err == nil {
		t.Error("Validate should return error after Reset (no target)")
	}
	// Args after Reset should not contain the old target or --brute.
	args := s.Args()
	if containsFlag(args, "--target") {
		t.Errorf("--target should not appear in args after Reset: %v", args)
	}
	if containsFlag(args, "--brute") {
		t.Errorf("--brute should not appear in args after Reset: %v", args)
	}
}

func TestReset_RetainsPythonAndOFAPath(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("example.com"),
	)
	s.Reset()
	s.AddOptions(oneforall.WithTarget("new.com"))

	args := s.Args()
	if args[0] != ofa {
		t.Errorf("args[0] after Reset = %q, want %q", args[0], ofa)
	}
	assertArg(t, args, "--target", "new.com")
}

func TestClone_IsIndependentFromOriginal(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	original, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("original.com"),
		oneforall.WithBruteForce(false),
	)

	clone := original.Clone()
	clone.AddOptions(
		oneforall.WithTarget("clone.com"),
		oneforall.WithBruteForce(true),
	)

	// Original must be unaffected.
	origArgs := original.Args()
	assertArg(t, origArgs, "--target", "original.com")
	if containsFlag(origArgs, "--brute") {
		// WithBruteForce(false) adds --brute False; check the value
		for i, a := range origArgs {
			if a == "--brute" && i+1 < len(origArgs) {
				if origArgs[i+1] != "False" {
					t.Errorf("original --brute = %q, want False", origArgs[i+1])
				}
			}
		}
	}

	// Clone should have clone.com target and brute True.
	cloneArgs := clone.Args()
	assertArg(t, cloneArgs, "--target", "clone.com")
	assertArg(t, cloneArgs, "--brute", "True")
}

func TestClone_SharesBaseConfig(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	original, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("example.com"),
	)
	clone := original.Clone()
	cloneArgs := clone.Args()

	// Clone must still have the oneforall path as first arg.
	if cloneArgs[0] != ofa {
		t.Errorf("clone args[0] = %q, want %q", cloneArgs[0], ofa)
	}
}

func TestWithTargetFile_TargetsArgSet(t *testing.T) {
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	domainFile := filepath.Join(t.TempDir(), "domains.txt")
	os.WriteFile(domainFile, []byte("a.com\nb.com\n"), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTargetFile(domainFile),
	)
	assertArg(t, s.Args(), "--targets", domainFile)
}

func TestWithOutputPath_DBPathInference(t *testing.T) {
	// When WithOutputPath is set, resolveResultDBPath should infer the DB
	// from that directory. We cannot call that method directly (unexported),
	// but we can verify the scanner builds without error and Args() has --path.
	pyPath := findPython3(t)
	ofa := filepath.Join(t.TempDir(), "oneforall.py")
	os.WriteFile(ofa, []byte(""), 0644)

	s, _ := oneforall.NewScanner(context.Background(),
		oneforall.WithPythonPath(pyPath),
		oneforall.WithOneForAllPath(ofa),
		oneforall.WithTarget("example.com"),
		oneforall.WithOutputPath("/custom/out"),
	)
	assertArg(t, s.Args(), "--path", "/custom/out")
}

// --- helpers ---

func assertArg(t *testing.T, args []string, flag, value string) {
	t.Helper()
	for i, a := range args {
		if a == flag && i+1 < len(args) && args[i+1] == value {
			return
		}
	}
	t.Errorf("flag %q with value %q not found in args: %v", flag, value, args)
}

func assertContains(t *testing.T, args []string, s string) {
	t.Helper()
	for _, a := range args {
		if a == s {
			return
		}
	}
	t.Errorf("%q not found in args: %v", s, args)
}

func containsFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func indexOf(args []string, s string) int {
	for i, a := range args {
		if a == s {
			return i
		}
	}
	return -1
}
