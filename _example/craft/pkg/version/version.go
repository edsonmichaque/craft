// Package version provides build and version information for the application.
// This information is populated at build time using -ldflags.
package version

import (
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// Build information, populated at build-time via -ldflags.
var (
	// Version indicates the current version of the application.
	// For releases, this will be a semantic version (e.g., "v1.2.3").
	// For development builds, this will be "dev" or a commit hash.
	Version = "dev"

	// GitCommit is the full git commit hash.
	GitCommit = "unknown"

	// GitBranch is the git branch from which the application was built.
	GitBranch = "unknown"

	// BuildTime is the UTC timestamp of when the binary was built.
	BuildTime = "unknown"

	// BuildUser is the username of who built the binary.
	BuildUser = "unknown"

	// GoVersion is the version of Go used to build the application.
	GoVersion = runtime.Version()

	// Platform is the target platform (OS/architecture combination).
	Platform = runtime.GOOS + "/" + runtime.GOARCH
)

// Info holds all version information.
type Info struct {
	Version      string            ` + "`json:\"version\"`" + `
	GitCommit    string            ` + "`json:\"gitCommit\"`" + `
	GitBranch    string            ` + "`json:\"gitBranch\"`" + `
	BuildTime    string            ` + "`json:\"buildTime\"`" + `
	BuildUser    string            ` + "`json:\"buildUser\"`" + `
	GoVersion    string            ` + "`json:\"goVersion\"`" + `
	Platform     string            ` + "`json:\"platform\"`" + `
	Dependencies map[string]string ` + "`json:\"dependencies,omitempty\"`" + `
}

// Get returns the version information as a structured object.
func Get() Info {
	return Info{
		Version:      Version,
		GitCommit:    GitCommit,
		GitBranch:    GitBranch,
		BuildTime:    BuildTime,
		BuildUser:    BuildUser,
		GoVersion:    GoVersion,
		Platform:     Platform,
		Dependencies: getDependencyVersions(),
	}
}

// String returns a human-readable version string.
func String() string {
	info := Get()
	return fmt.Sprintf(
		"Version:      %s\n"+
			"Git Commit:   %s\n"+
			"Git Branch:   %s\n"+
			"Built:        %s\n"+
			"Built By:     %s\n"+
			"Go Version:   %s\n"+
			"Platform:     %s",
		info.Version,
		info.GitCommit,
		info.GitBranch,
		info.BuildTime,
		info.BuildUser,
		info.GoVersion,
		info.Platform,
	)
}

// JSON returns version information in JSON format.
func JSON(indent bool) string {
	info := Get()
	var data []byte
	var err error

	if indent {
		data, err = json.MarshalIndent(info, "", "  ")
	} else {
		data, err = json.Marshal(info)
	}

	if err != nil {
		return fmt.Sprintf("{\n  \"error\": \"%s\"\n}", err)
	}
	return string(data)
}

// Short returns a condensed version string.
func Short() string {
	commit := GitCommit
	if len(commit) > 8 {
		commit = commit[:8]
	}
	if Version == "dev" {
		return fmt.Sprintf("dev+%s", commit)
	}
	return fmt.Sprintf("%s+%s", Version, commit)
}

// IsRelease returns true if the current version represents a release build.
func IsRelease() bool {
	return Version != "dev" && strings.HasPrefix(Version, "v")
}

// IsDevelopment returns true if this is a development build.
func IsDevelopment() bool {
	return Version == "dev"
}

// getDependencyVersions returns a map of direct dependencies and their versions.
func getDependencyVersions() map[string]string {
	deps := make(map[string]string)
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return deps
	}

	for _, dep := range info.Deps {
		// Only include direct dependencies
		if !strings.Contains(dep.Path, "/internal/") {
			deps[dep.Path] = dep.Version
		}
	}
	return deps
}

// BuildContext returns a map of build-time variables.
func BuildContext() map[string]string {
	return map[string]string{
		"version":    Version,
		"gitCommit": GitCommit,
		"gitBranch": GitBranch,
		"buildTime": BuildTime,
		"buildUser": BuildUser,
		"goVersion": GoVersion,
		"platform":  Platform,
	}
}

// Validate checks if the version information appears to be properly populated.
func Validate() error {
	if Version == "dev" && GitCommit == "unknown" {
		return fmt.Errorf("version information not properly initialized")
	}
	return nil
}