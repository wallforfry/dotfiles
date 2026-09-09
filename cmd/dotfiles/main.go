package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wallforfry/dotfiles/internal/audit"
	"github.com/wallforfry/dotfiles/internal/commands"
	"github.com/wallforfry/dotfiles/internal/verify"
)

var directCommands = map[string]func(commands.Runtime, []string) int{
	"agent-handoff": commands.AgentHandoff,
	"cloak":         commands.Cloak,
	"firecrawl-mcp": commands.Firecrawl,
	"postgres-mcp":  commands.Postgres,
	"scrapling-mcp": commands.Scrapling,
}

func main() {
	os.Exit(run(os.Args, commands.NewRuntime()))
}

func run(arguments []string, runtime commands.Runtime) int {
	name := filepath.Base(arguments[0])
	args := arguments[1:]
	if command, exists := directCommands[name]; exists {
		return command(runtime, args)
	}
	if len(args) == 0 {
		usage(runtime)
		return commands.ExitUsage
	}
	name, args = args[0], args[1:]
	if command, exists := directCommands[name]; exists {
		return command(runtime, args)
	}
	switch name {
	case "register-claude-hook":
		return commands.RegisterClaudeHook(runtime, args)
	case "harness-audit":
		root, code := requiredRepositoryRoot(args, runtime)
		if code != 0 {
			return code
		}
		return audit.Run(auditConfig(root), runtime.Stdout, runtime.Stderr)
	case "verify":
		root, code := repositoryRoot(args, runtime)
		if code != 0 {
			return code
		}
		return verify.Run(root, runtime.Stdout, runtime.Stderr)
	case "validate-skill-routing":
		if len(args) > 1 {
			fmt.Fprintln(runtime.Stderr, "dotfiles: usage - validate-skill-routing [corpus.tsv]")
			return commands.ExitUsage
		}
		root, code := repositoryRoot(nil, runtime)
		if code != 0 {
			return code
		}
		corpus := ""
		if len(args) == 1 {
			corpus = args[0]
		}
		return verify.RunRouting(root, corpus, runtime.Stdout, runtime.Stderr)
	default:
		fmt.Fprintf(runtime.Stderr, "dotfiles: commande inconnue « %s »\n", name)
		usage(runtime)
		return commands.ExitUsage
	}
}

func auditConfig(root string) audit.Config {
	return audit.Config{
		Root:           root,
		ClaudeProjects: os.Getenv("CLAUDE_PROJECTS"),
		CodexSessions:  os.Getenv("CODEX_SESSIONS"),
		Since:          os.Getenv("HARNESS_RULES_SINCE"),
	}
}

func requiredRepositoryRoot(args []string, runtime commands.Runtime) (string, int) {
	if len(args) == 2 && args[0] == "--repository" && args[1] != "" {
		return args[1], 0
	}
	fmt.Fprintln(runtime.Stderr, "dotfiles: usage - harness-audit --repository <chemin>")
	return "", commands.ExitUsage
}

func repositoryRoot(args []string, runtime commands.Runtime) (string, int) {
	if len(args) == 2 && args[0] == "--repository" && args[1] != "" {
		return args[1], 0
	}
	if len(args) != 0 {
		fmt.Fprintln(runtime.Stderr, "dotfiles: usage - verify [--repository <chemin>]")
		return "", commands.ExitUsage
	}
	command := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := command.Output()
	if err != nil {
		fmt.Fprintln(runtime.Stderr, "dotfiles: dépôt git introuvable, utiliser --repository")
		return "", commands.ExitUsage
	}
	return strings.TrimSpace(string(output)), 0
}

func usage(runtime commands.Runtime) {
	fmt.Fprintln(runtime.Stderr, "dotfiles: commandes - verify, harness-audit, validate-skill-routing, register-claude-hook, agent-handoff, cloak, firecrawl-mcp, postgres-mcp, scrapling-mcp")
}
