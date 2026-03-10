package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProjectConfig represents the .polaris.yaml project configuration file
type ProjectConfig struct {
	Project    string             `yaml:"project"`
	Components []ProjectComponent `yaml:"components"`
}

// ProjectComponent represents a component within a project
type ProjectComponent struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

// LoadProjectConfigFrom reads .polaris.yaml from the specified directory's git root.
// If targetDir is empty, uses the current working directory (existing behavior).
func LoadProjectConfigFrom(targetDir string) *ProjectConfig {
	var gitRoot string
	if targetDir != "" {
		absTarget, err := filepath.Abs(targetDir)
		if err != nil {
			return nil
		}
		cmd := exec.Command("git", "-C", absTarget, "rev-parse", "--show-toplevel")
		out, err := cmd.Output()
		if err != nil {
			// Not a git repo — try using the directory itself
			gitRoot = absTarget
		} else {
			gitRoot = strings.TrimSpace(string(out))
		}
	} else {
		cmd := exec.Command("git", "rev-parse", "--show-toplevel")
		out, err := cmd.Output()
		if err != nil {
			return nil
		}
		gitRoot = strings.TrimSpace(string(out))
	}

	data, err := os.ReadFile(filepath.Join(gitRoot, ".polaris.yaml"))
	if err != nil {
		return nil
	}

	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	return &cfg
}

// LoadProjectConfig reads .polaris.yaml from the current directory's git root.
func LoadProjectConfig() *ProjectConfig {
	return LoadProjectConfigFrom("")
}

// WriteProjectConfig writes a ProjectConfig to disk as .polaris.yaml
func WriteProjectConfig(path string, cfg *ProjectConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	header := "# Polaris project configuration\n# Used by detect-risks and reliability-review skills for consistent service naming\n"
	return os.WriteFile(path, []byte(header+string(data)), 0644)
}
