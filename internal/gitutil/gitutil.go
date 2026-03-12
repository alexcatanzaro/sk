package gitutil

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// minMajor and minMinor define the minimum required git version.
const (
	minMajor = 2
	minMinor = 25
)

// CheckVersion verifies that the installed git is at least 2.25.
func CheckVersion() error {
	out, err := exec.Command("git", "--version").Output()
	if err != nil {
		return fmt.Errorf("git not found: %w\n  → install git and ensure it is on your PATH", err)
	}
	// output: "git version 2.39.3\n"
	line := strings.TrimSpace(string(out))
	parts := strings.Fields(line) // ["git", "version", "2.39.3"]
	if len(parts) < 3 {
		return fmt.Errorf("could not parse git version output: %q", line)
	}
	ver := parts[2]
	segments := strings.SplitN(ver, ".", 3)
	if len(segments) < 2 {
		return fmt.Errorf("unexpected git version format: %q", ver)
	}
	major, err := strconv.Atoi(segments[0])
	if err != nil {
		return fmt.Errorf("unexpected git version format: %q", ver)
	}
	minor, err := strconv.Atoi(segments[1])
	if err != nil {
		return fmt.Errorf("unexpected git version format: %q", ver)
	}
	if major < minMajor || (major == minMajor && minor < minMinor) {
		return fmt.Errorf("sk requires git ≥ %d.%d for sparse checkout support (found %s)\n  → upgrade git and retry", minMajor, minMinor, ver)
	}
	return nil
}

// RunIn runs the git command with args inside dir, streaming stderr to the
// caller and returning a combined error on failure.
func RunIn(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(buf.String()))
	}
	return nil
}

// OutputIn runs a git command in dir and returns its stdout.
func OutputIn(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(string(ee.Stderr)))
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

// InitSparseCheckout initialises a git sparse checkout in dir, fetching only
// subpath from the remote url.
func InitSparseCheckout(dir, url, subpath string) error {
	if err := CheckVersion(); err != nil {
		return err
	}
	if err := RunIn(dir, "init"); err != nil {
		return err
	}
	if err := RunIn(dir, "remote", "add", "origin", url); err != nil {
		return err
	}
	// Enable sparse checkout and restrict to the skills subpath.
	if err := RunIn(dir, "sparse-checkout", "set", "--no-cone", subpath+"/**"); err != nil {
		// Fallback for older cone-only implementations
		if err2 := RunIn(dir, "sparse-checkout", "set", subpath); err2 != nil {
			return err
		}
	}
	if err := RunIn(dir, "fetch", "--depth=1", "origin"); err != nil {
		return err
	}
	// Checkout the fetched HEAD.
	if err := RunIn(dir, "checkout", "FETCH_HEAD"); err != nil {
		// Try checking out origin/HEAD or main as fallback.
		if err2 := RunIn(dir, "checkout", "-b", "main", "origin/HEAD"); err2 != nil {
			return err
		}
	}
	return nil
}

// Refresh pulls the latest changes in an existing sparse checkout.
func Refresh(dir string) error {
	if err := CheckVersion(); err != nil {
		return err
	}
	if err := RunIn(dir, "fetch", "--depth=1", "origin"); err != nil {
		return err
	}
	return RunIn(dir, "checkout", "FETCH_HEAD")
}
