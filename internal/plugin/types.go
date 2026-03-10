package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// PluginInfo tracks installed plugin metadata
type PluginInfo struct {
	Editor    string `json:"editor"`
	Version   string `json:"version"`
	Installed string `json:"installed"` // ISO8601 timestamp
	Location  string `json:"location"`
}

// PolarisSkillNames lists all Polaris skill directory names for cleanup
var PolarisSkillNames = []string{
	"detect-risks", "analyze-risks", "remediate-risks", "risk-check",
	"risk-guidance", "control-guidance", "submit-evidence", "reliability-review",
	"incident-patterns", "sre-context", "list-open",
}

// EditorBinaries maps editor names to their CLI binary names for PATH detection
var EditorBinaries = []struct {
	Name   string
	Binary string
}{
	{"claude", "claude"},
	{"codex", "codex"},
	{"gemini", "gemini"},
	{"cursor", "cursor"},
	{"windsurf", "windsurf"},
	{"copilot", "copilot"},
	{"augment", "auggie"},
}

// SavePluginInfo persists plugin metadata to ~/.polaris/plugins.json
func SavePluginInfo(editor, version, location string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	metadataDir := filepath.Join(home, ".polaris")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return err
	}

	metadataFile := filepath.Join(metadataDir, "plugins.json")

	// Load existing metadata
	var plugins []PluginInfo
	if data, err := os.ReadFile(metadataFile); err == nil {
		_ = json.Unmarshal(data, &plugins)
	}

	// Update or add this plugin
	found := false
	for i, p := range plugins {
		if p.Editor == editor {
			plugins[i].Version = version
			plugins[i].Installed = time.Now().Format(time.RFC3339)
			plugins[i].Location = location
			found = true
			break
		}
	}

	if !found {
		plugins = append(plugins, PluginInfo{
			Editor:    editor,
			Version:   version,
			Installed: time.Now().Format(time.RFC3339),
			Location:  location,
		})
	}

	// Save metadata
	data, err := json.MarshalIndent(plugins, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataFile, data, 0644)
}

// GetInstalledPlugins reads the plugin metadata file
func GetInstalledPlugins() ([]PluginInfo, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	metadataFile := filepath.Join(home, ".polaris", "plugins.json")
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []PluginInfo{}, nil
		}
		return nil, err
	}

	var plugins []PluginInfo
	if err := json.Unmarshal(data, &plugins); err != nil {
		return nil, err
	}

	return plugins, nil
}

// RemovePluginFromMetadata removes a plugin entry from the metadata file
func RemovePluginFromMetadata(editor, metadataFile string) error {
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return err
	}

	var plugins []PluginInfo
	if err := json.Unmarshal(data, &plugins); err != nil {
		return err
	}

	// Filter out the removed plugin
	var updated []PluginInfo
	for _, p := range plugins {
		if p.Editor != editor {
			updated = append(updated, p)
		}
	}

	data, err = json.MarshalIndent(updated, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataFile, data, 0644)
}
