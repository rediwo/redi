package runtime

import (
	"os/exec"
	"strings"
)

// VersionProvider provides version information
type VersionProvider struct {
	buildVersion string
}

// NewVersionProvider creates a new version provider
func NewVersionProvider(buildVersion string) *VersionProvider {
	return &VersionProvider{
		buildVersion: buildVersion,
	}
}

// GetVersion returns the runtime version
func (v *VersionProvider) GetVersion() string {
	// If version was set at build time, use it
	if v.buildVersion != "" && v.buildVersion != "dev" {
		return v.buildVersion
	}
	
	// Try to get version from git tag
	if gitVersion := v.getGitVersion(); gitVersion != "" {
		return gitVersion
	}
	
	// Fallback to dev version
	return "dev"
}

// getGitVersion attempts to get the current version from git tags
func (v *VersionProvider) getGitVersion() string {
	cmd := exec.Command("git", "describe", "--tags", "--exact-match", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		// Try to get the latest tag with commit info
		cmd = exec.Command("git", "describe", "--tags", "--always")
		if output, err = cmd.Output(); err != nil {
			return ""
		}
	}
	
	version := strings.TrimSpace(string(output))
	if version == "" {
		return ""
	}
	
	return version
}

// DefaultVersionProvider with "dev" build version
var DefaultVersionProvider = NewVersionProvider("dev")