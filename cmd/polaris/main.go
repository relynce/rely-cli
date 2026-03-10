// Package main provides the polaris CLI for secure interaction with Polaris API.
// This CLI acts as a trusted intermediary - credentials are stored locally and
// never exposed to LLM contexts.
package main

import (
	"fmt"
	"os"

	"github.com/relynce/polaris-cli/internal/commands"
	"github.com/relynce/polaris-cli/internal/plugin"
)

// version and gitHash are set at build time via -ldflags "-X main.version=... -X main.gitHash=..."
var version = "source-build"
var gitHash = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "init":
		commands.CmdInit(os.Args[2:])
	case "login":
		commands.CmdLogin()
	case "logout":
		commands.CmdLogout()
	case "status":
		commands.CmdStatus(version, gitHash)
	case "scan":
		commands.CmdScan(os.Args[2:], version)
	case "risk":
		commands.CmdRisk(os.Args[2:])
	case "control":
		commands.CmdControl(os.Args[2:])
	case "knowledge":
		commands.CmdKnowledge(os.Args[2:])
	case "evidence":
		commands.CmdEvidence(os.Args[2:])
	case "config":
		commands.CmdConfig(os.Args[2:])
	case "plugin":
		plugin.CmdPlugin(os.Args[2:])
	case "version":
		fmt.Printf("polaris version %s (%s)\n", version, gitHash)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`polaris - Secure CLI for Polaris reliability analysis

Usage:
  polaris <command> [options]

Commands:
  init               Initialize Polaris for this repository
  login              Configure credentials interactively
  logout             Remove stored credentials
  status             Check connection and authentication status
  scan               Submit risk findings to Polaris
  risk               Manage risk lifecycle (list, close, resolve, etc.)
  control            Query reliability controls catalog
  knowledge          Query organizational knowledge base (facts, procedures, patterns)
  evidence           Manage control evidence (submit, list, verify)
  plugin             Manage editor plugins (install, update, list, remove)
  config show        Show current configuration (API key masked)
  config set <k> <v> Set a configuration value
  version            Show version information
  help               Show this help message

Scan Command:
  polaris scan --service <name> --stdin       Read findings JSON from stdin
  polaris scan --service <name> --file <path> Read findings from file
  polaris scan --service <name> --dry-run     Validate without submitting
  polaris scan --target <path> --file <path>  Scan another project (service auto-resolved from .polaris.yaml)

Risk Command:
  polaris risk list [--status=detected] [--service=name]  List risks
  polaris risk show <risk-code>                           Show risk details with mapped controls
  polaris risk stale [--service=name]                     List stale risks
  polaris risk close <risk-code> [--reason="..."]         Close a risk
  polaris risk resolve <risk-code> --reason="..."         Mark risk as resolved
  polaris risk acknowledge <risk-code> [<risk-code>...]   Acknowledge risks
  polaris risk accept <risk-code> --reason="..."          Accept risk (won't mitigate)

Control Command:
  polaris control list [--category=<cat>]     List controls in catalog
  polaris control show <control-code>         Show control details (e.g., RC-018)

Examples:
  # Initial setup
  polaris login

  # Submit findings from Claude Code skill
  echo '{"findings":[...]}' | polaris scan --service checkout-api --stdin

  # Scan a different project (service name auto-resolved from target's .polaris.yaml)
  polaris scan --target /path/to/other-project --file findings.json

  # Check status
  polaris status

  # Manage risks
  polaris risk list --status=detected
  polaris risk close R-001 --reason "Fixed by implementing timeout"
  polaris risk stale --service checkout-api

  # Query controls catalog
  polaris control list --category=fault_tolerance
  polaris control show RC-018

  # Query knowledge base
  polaris knowledge search "circuit breaker timeout"
  polaris knowledge procedures --control=RC-018
  polaris knowledge patterns --type=failure_mode

  # Submit evidence for controls
  polaris evidence submit --control=RC-018 --type=code --name="Circuit breaker impl" --url="https://github.com/..."
  polaris evidence list --status=configured
  polaris evidence verify <evidence-id>

Plugin Command:
  polaris plugin install <editor>      Install plugin for editor (claude, codex, gemini, cursor, windsurf, copilot, augment)
  polaris plugin update [editor]       Update plugin(s) to latest version
  polaris plugin list                  List installed plugins
  polaris plugin remove <editor>       Remove installed plugin

Init Command:
  polaris init                         Interactive initialization
  polaris init --project <name>        Set project name non-interactively
  polaris init --skip-plugin           Skip plugin installation
  polaris init --force                 Overwrite existing config without prompting
  polaris init -y                      Accept all defaults

Configuration:
  Credentials are stored in ~/.polaris/config.yaml
  Never share this file or expose credentials to LLM contexts.`)
}
