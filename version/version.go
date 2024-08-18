package version

import (
	"fmt"
	"os/exec"
	"strings"
)

func ExecuteGitCommand(args ...string) string {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

var (
	// Version of the release, dynamically fetched from git
	version = ExecuteGitCommand("describe", "--tags", "--abbrev=0")

	// Commit hash of the release, dynamically fetched from git
	commit = ExecuteGitCommand("rev-parse", "HEAD")

	// Date of the commit, dynamically fetched from git
	date = ExecuteGitCommand("show", "-s", "--format=%ci", "HEAD")
)

func GetVersion() string {
	return version
}

func GetFormattedVersion() string {
	return fmt.Sprintf("%s (%s, %s)", version, commit, date)
}
