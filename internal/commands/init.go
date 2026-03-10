package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/relynce/polaris-cli/internal/api"
	"github.com/relynce/polaris-cli/internal/config"
	"github.com/relynce/polaris-cli/internal/plugin"
	"github.com/relynce/polaris-cli/internal/project"
	"gopkg.in/yaml.v3"
)

const agentsMdTemplate = `## Polaris

This project uses Polaris for reliability risk analysis. The following skills are available:

### Risk Detection
- ` + "`/polaris:detect-risks`" + ` — Scan code for reliability risks and submit findings
- ` + "`/polaris:risk-guidance`" + ` — Get detailed guidance for a specific risk

### Risk Remediation
- ` + "`/polaris:remediate-risks`" + ` — Auto-implement fixes for detected risks

### Quick Reference
- Run ` + "`polaris risk list`" + ` to see current risks
- Run ` + "`polaris risk show <code>`" + ` for risk details with mapped controls
- Run ` + "`polaris control show <code>`" + ` for control implementation guidance
`

func printInitUsage() {
	fmt.Println(`polaris init - Initialize Polaris for this repository

Usage:
  polaris init [options]

Options:
  --project <name>    Set project name (default: from git remote or directory name)
  --skip-plugin       Skip installing the Polaris plugin for Claude Code
  --force             Overwrite existing config and plugin without prompting
  -y, --yes           Accept all defaults non-interactively

What it does:
  1. Creates .polaris.yaml with project name and detected components
  2. Installs the Polaris plugin for Claude Code (if available)
  3. Adds Polaris sections to AGENTS.md (creates or appends)
  4. Checks if API credentials are configured

Examples:
  polaris init                         Interactive setup
  polaris init --project my-service    Set project name directly
  polaris init -y                      Accept all auto-detected defaults
  polaris init --force                 Overwrite existing config`)
}

// CmdInit initializes Polaris for a repository
func CmdInit(args []string) {
	var projectName string
	var skipPlugin bool
	var force bool
	var yesAll bool

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "help", "--help", "-h":
			printInitUsage()
			return
		case "--skip-plugin", "--skip-skills":
			skipPlugin = true
		case "--force":
			force = true
		case "-y", "--yes":
			yesAll = true
		case "--project":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "Error: --project requires a value")
				os.Exit(1)
			}
			i++
			projectName = args[i]
		default:
			if strings.HasPrefix(args[i], "--project=") {
				projectName = strings.TrimPrefix(args[i], "--project=")
			}
		}
	}

	fmt.Println("Initializing Polaris...")
	fmt.Println()

	// Step 1: Require git repo
	gitRoot := project.DetectGitRoot()
	if gitRoot == "" {
		fmt.Fprintln(os.Stderr, "Error: not a git repository.")
		fmt.Fprintln(os.Stderr, "Polaris must be initialized inside a git repository.")
		fmt.Fprintln(os.Stderr, "Run 'git init' first, then try again.")
		os.Exit(1)
	}

	// Step 2: Generate .polaris.yaml
	configPath := filepath.Join(gitRoot, ".polaris.yaml")
	writeConfig := true

	if _, err := os.Stat(configPath); err == nil {
		if yesAll {
			writeConfig = false
			fmt.Println("Keeping existing .polaris.yaml (use interactive mode to overwrite)")
		} else {
			existing, _ := os.ReadFile(configPath)
			fmt.Println("Existing .polaris.yaml found:")
			fmt.Println(string(existing))

			var overwrite bool
			err := huh.NewConfirm().
				Title("Overwrite existing .polaris.yaml?").
				Affirmative("Yes").
				Negative("No").
				Value(&overwrite).
				Run()
			if err != nil || !overwrite {
				writeConfig = false
				fmt.Println("Keeping existing .polaris.yaml")
			}
		}
	}

	var cfg *project.ProjectConfig
	if writeConfig {
		cfg = buildProjectConfig(gitRoot, projectName, yesAll)
		if err := project.WriteProjectConfig(configPath, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing .polaris.yaml: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created .polaris.yaml (project: %s, %d components)\n", cfg.Project, len(cfg.Components))
	} else {
		// Load existing config for summary
		data, _ := os.ReadFile(configPath)
		cfg = &project.ProjectConfig{}
		_ = yaml.Unmarshal(data, cfg)
	}
	fmt.Println()

	// Step 3: Install skills for detected editors
	pluginInstalled := false
	pluginVersion := ""
	if !skipPlugin {
		plugins, _ := plugin.GetInstalledPlugins()
		installedMap := make(map[string]*plugin.PluginInfo)
		for i := range plugins {
			installedMap[plugins[i].Editor] = &plugins[i]
		}

		// Fetch server version once for update checks
		loginCfgForPlugin, _ := config.LoadConfig()
		serverVersion := api.FetchServerPluginVersion(loginCfgForPlugin)

		// Detect available editors
		var detectedEditors []string
		for _, e := range plugin.EditorBinaries {
			if plugin.IsEditorAvailable(e.Binary) {
				detectedEditors = append(detectedEditors, e.Name)
			}
		}

		if len(detectedEditors) == 0 {
			fmt.Println("Skills: No supported editors detected on PATH")
			fmt.Println("  Supported: claude, codex, gemini, cursor, windsurf, copilot, augment")
			fmt.Println("  Install an editor, then run: polaris plugin install <editor>")
		}

		for _, editorName := range detectedEditors {
			existing := installedMap[editorName]

			if existing != nil {
				// Already installed — check for updates
				if serverVersion != "" && serverVersion != existing.Version {
					doUpdate := force || yesAll
					if !doUpdate {
						err := huh.NewConfirm().
							Title(fmt.Sprintf("Update Polaris skills for %s? (v%s → v%s)", editorName, existing.Version, serverVersion)).
							Affirmative("Yes").
							Negative("No").
							Value(&doUpdate).
							Run()
						if err != nil {
							doUpdate = false
						}
					}
					if doUpdate {
						if err := plugin.InstallPlugin(editorName); err != nil {
							fmt.Fprintf(os.Stderr, "Warning: could not update %s skills: %v\n", editorName, err)
						} else {
							pluginInstalled = true
							pluginVersion = serverVersion
						}
					} else {
						pluginInstalled = true
						pluginVersion = existing.Version
						fmt.Printf("Skills (%s): Keeping v%s\n", editorName, existing.Version)
					}
				} else {
					pluginInstalled = true
					pluginVersion = existing.Version
					fmt.Printf("Skills (%s): Up to date (v%s)\n", editorName, existing.Version)
				}
			} else {
				// Editor available but skills not installed
				doInstall := yesAll
				if !yesAll {
					err := huh.NewConfirm().
						Title(fmt.Sprintf("Install Polaris skills for %s?", editorName)).
						Affirmative("Yes").
						Negative("No").
						Value(&doInstall).
						Run()
					if err != nil {
						doInstall = false
					}
				}
				if doInstall {
					if err := plugin.InstallPlugin(editorName); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: could not install %s skills: %v\n", editorName, err)
					} else {
						pluginInstalled = true
						// Read back the version from metadata
						updatedPlugins, _ := plugin.GetInstalledPlugins()
						for _, p := range updatedPlugins {
							if p.Editor == editorName {
								pluginVersion = p.Version
								break
							}
						}
					}
				}
			}
		}
		fmt.Println()
	}

	// Step 4: Set up AGENTS.md
	agentsMdAction := ""
	action, err := EnsureAgentsMd(gitRoot, force, yesAll)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not set up AGENTS.md: %v\n", err)
	} else {
		agentsMdAction = action
		switch action {
		case "created":
			fmt.Println("Created AGENTS.md with Polaris sections")
		case "appended":
			fmt.Println("Appended Polaris sections to AGENTS.md")
		case "updated":
			fmt.Println("Updated Polaris sections in AGENTS.md")
		case "skipped":
			fmt.Println("AGENTS.md: Skipped")
		}
	}
	fmt.Println()

	// Step 5: Check credentials
	credentialsConfigured := false
	credentialsURL := ""
	loginCfg, _ := config.LoadConfig()
	if loginCfg != nil && loginCfg.APIKey != "" {
		credentialsConfigured = true
		credentialsURL = loginCfg.APIURL
		fmt.Printf("Credentials: Configured (API URL: %s)\n", credentialsURL)
	} else {
		fmt.Println("Credentials: Not configured")
		fmt.Println("  Run 'polaris login' to set up API credentials.")
	}
	fmt.Println()

	// Step 6: Print summary
	printInitSummary(cfg, pluginInstalled, pluginVersion, credentialsConfigured, agentsMdAction)
}

// buildProjectConfig creates a ProjectConfig interactively or from defaults
func buildProjectConfig(gitRoot, projectName string, yesAll bool) *project.ProjectConfig {
	// Auto-detect project name
	if projectName == "" {
		projectName = project.DetectProjectName(gitRoot)
	}

	// Auto-detect components
	components := project.DetectComponents(gitRoot)

	if yesAll {
		if len(components) == 0 {
			components = []project.ProjectComponent{{Name: projectName, Path: "."}}
		}
		return &project.ProjectConfig{Project: projectName, Components: components}
	}

	// Interactive: prompt for project name
	err := huh.NewInput().
		Title("Project name").
		Value(&projectName).
		Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Show detected components
	if len(components) > 0 {
		fmt.Println("Detected components:")
		for i, c := range components {
			fmt.Printf("  %d. %-20s %s\n", i+1, c.Name, c.Path)
		}
		fmt.Println()

		var accept bool
		err := huh.NewConfirm().
			Title("Accept detected components?").
			Affirmative("Yes").
			Negative("No, let me edit").
			Value(&accept).
			Run()
		if err != nil {
			os.Exit(1)
		}

		if !accept {
			components = promptComponents()
		}
	} else {
		fmt.Println("No components auto-detected.")
		fmt.Println()

		var addManual bool
		err := huh.NewConfirm().
			Title("Add components manually?").
			Affirmative("Yes").
			Negative("No, use project root").
			Value(&addManual).
			Run()
		if err != nil {
			os.Exit(1)
		}

		if addManual {
			components = promptComponents()
		} else {
			components = []project.ProjectComponent{{Name: projectName, Path: "."}}
		}
	}

	return &project.ProjectConfig{Project: projectName, Components: components}
}

// promptComponents interactively collects component definitions
func promptComponents() []project.ProjectComponent {
	var components []project.ProjectComponent
	for {
		var name, path string

		err := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Component name").
					Value(&name),
				huh.NewInput().
					Title("Component path (relative to repo root)").
					Value(&path),
			),
		).Run()
		if err != nil {
			break
		}

		if name == "" || path == "" {
			break
		}

		// Ensure path ends with /
		if path != "." && !strings.HasSuffix(path, "/") {
			path += "/"
		}

		components = append(components, project.ProjectComponent{Name: name, Path: path})
		fmt.Printf("  Added: %s -> %s\n", name, path)

		var addMore bool
		err = huh.NewConfirm().
			Title("Add another component?").
			Affirmative("Yes").
			Negative("Done").
			Value(&addMore).
			Run()
		if err != nil || !addMore {
			break
		}
	}
	return components
}

// EnsureAgentsMd creates or updates AGENTS.md with Polaris sections
func EnsureAgentsMd(gitRoot string, force, yesAll bool) (string, error) {
	agentsMdPath := filepath.Join(gitRoot, "AGENTS.md")
	content, err := os.ReadFile(agentsMdPath)

	if os.IsNotExist(err) {
		if err := os.WriteFile(agentsMdPath, []byte(agentsMdTemplate), 0644); err != nil {
			return "", err
		}
		return "created", nil
	}

	if err != nil {
		return "", err
	}

	contentStr := string(content)
	hasPolarisSection := strings.Contains(contentStr, "## Polaris")

	if !hasPolarisSection {
		var shouldAppend bool
		if yesAll || force {
			shouldAppend = true
		} else {
			err := huh.NewConfirm().
				Title("AGENTS.md exists but has no Polaris section. Append?").
				Affirmative("Yes").
				Negative("No").
				Value(&shouldAppend).
				Run()
			if err != nil {
				return "skipped", nil
			}
		}

		if shouldAppend {
			updatedContent := contentStr
			if !strings.HasSuffix(contentStr, "\n") {
				updatedContent += "\n"
			}
			updatedContent += "\n" + agentsMdTemplate
			if err := os.WriteFile(agentsMdPath, []byte(updatedContent), 0644); err != nil {
				return "", err
			}
			return "appended", nil
		}
		return "skipped", nil
	}

	// Already has Polaris section — prompt to update
	var shouldUpdate bool
	if yesAll || force {
		shouldUpdate = true
	} else {
		err := huh.NewConfirm().
			Title("AGENTS.md already has Polaris section. Update?").
			Affirmative("Yes").
			Negative("No").
			Value(&shouldUpdate).
			Run()
		if err != nil {
			return "skipped", nil
		}
	}

	if shouldUpdate {
		lines := strings.Split(contentStr, "\n")
		var newLines []string
		var inPolarisSection bool

		for i, line := range lines {
			if strings.TrimSpace(line) == "## Polaris" {
				inPolarisSection = true
				newLines = append(newLines, agentsMdTemplate)
				continue
			}

			if inPolarisSection {
				if strings.HasPrefix(strings.TrimSpace(line), "##") && line != "## Polaris" {
					inPolarisSection = false
					newLines = append(newLines, line)
				}
				if i == len(lines)-1 {
					break
				}
				continue
			}

			newLines = append(newLines, line)
		}

		updatedContent := strings.Join(newLines, "\n")
		if err := os.WriteFile(agentsMdPath, []byte(updatedContent), 0644); err != nil {
			return "", err
		}
		return "updated", nil
	}

	return "skipped", nil
}

func printInitSummary(cfg *project.ProjectConfig, pluginInstalled bool, pluginVersion string, credentialsConfigured bool, agentsMdAction string) {
	fmt.Println("=== Polaris Initialization Complete ===")
	fmt.Println()
	fmt.Printf("Project: %s\n", cfg.Project)
	fmt.Printf("Components: %d\n", len(cfg.Components))
	for _, c := range cfg.Components {
		fmt.Printf("  - %s (%s)\n", c.Name, c.Path)
	}
	fmt.Println()
	if pluginInstalled {
		fmt.Printf("Skills: Installed (v%s)\n", pluginVersion)
	} else {
		fmt.Println("Skills: Not installed")
	}
	if agentsMdAction != "" && agentsMdAction != "skipped" {
		fmt.Printf("AGENTS.md: %s\n", agentsMdAction)
	}
	if credentialsConfigured {
		fmt.Println("Credentials: Configured")
	} else {
		fmt.Println("Credentials: Not configured — run 'polaris login'")
	}
	fmt.Println()
	fmt.Println("Next steps:")
	if !credentialsConfigured {
		fmt.Println("  1. polaris login")
		fmt.Println("  2. polaris plugin install claude")
	} else if !pluginInstalled {
		fmt.Println("  1. polaris plugin install claude")
	}
	fmt.Println("  - Commit .polaris.yaml and AGENTS.md to your repository")
	fmt.Println("  - Use /polaris:detect-risks to scan for reliability risks")
}
