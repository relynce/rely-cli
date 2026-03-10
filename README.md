# Relynce CLI

Connect your codebase to the [Relynce](https://dev.relynce.ai) reliability risk platform. Scan for risks, get control guidance, and manage your reliability posture — all from the terminal or Claude Code.

## Install

**From source (requires Go 1.25+):**

```bash
go install github.com/relynce/rely-cli/cmd/rely@latest
```

**From release binary:**

Download from [Releases](https://github.com/relynce/rely-cli/releases) for your platform.

## Quick Start

```bash
# Configure your API credentials
rely login

# Initialize a project (installs Claude Code plugin if available)
rely init

# Check connection and plugin status
rely status
```

## Claude Code Integration

After `rely init` (or `rely plugin install claude`), the following slash commands are available in Claude Code. The core workflow chains automatically: detect → analyze → remediate.

**Multi-Agent Workflow:**

| Command | Description |
|---------|-------------|
| `/rely:detect-risks` | Multi-agent scan with expert agents, auto-chains to analyze |
| `/rely:analyze-risks` | Correlate with incidents, enrich with knowledge, score risks (auto-invoked) |
| `/rely:remediate-risks R-XXX` | Generate plan, apply fixes, submit evidence, resolve risk |

**Guidance and Research:**

| Command | Description |
|---------|-------------|
| `/rely:risk-guidance R-XXX` | Codebase-specific remediation guidance for a risk |
| `/rely:risk-check` | Quick read-only check of existing risks |
| `/rely:control-guidance RC-XXX` | Implementation guidance for a control |
| `/rely:incident-patterns` | Search historical incident patterns |
| `/rely:sre-context` | Load full reliability context |

**Review and Evidence:**

| Command | Description |
|---------|-------------|
| `/rely:reliability-review` | Review code changes for reliability |
| `/rely:submit-evidence` | Submit control implementation evidence |
| `/rely:list-open` | List unresolved risks |

## Commands

| Command | Description |
|---------|-------------|
| `rely login` | Configure API credentials |
| `rely logout` | Remove stored credentials |
| `rely init` | Initialize project and install Claude Code plugin |
| `rely status` | Check connection and plugin status |
| `rely plugin` | Manage editor plugins (install, update, list, remove) |
| `rely scan` | Submit risk scan findings |
| `rely risk` | Manage risks (list, show, close, resolve) |
| `rely control` | Query the 56-control reliability catalog |
| `rely knowledge` | Search organizational knowledge base |
| `rely evidence` | Submit and manage control evidence |
| `rely config` | Manage configuration |
| `rely version` | Show version info |

## Configuration

Credentials are stored in `~/.relynce/config.yaml` (mode 0600). The CLI never exposes credentials to LLM contexts.

## License

[Business Source License 1.1](LICENSE) — see LICENSE for details.
