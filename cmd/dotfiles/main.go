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

func main() {
	os.Exit(run(os.Args, commands.NewRuntime()))
}

// run route sur le catalogue, que le CLI soit appelé par son nom ou par un lien
// de ~/.local/bin portant le nom d'une commande.
func run(arguments []string, runtime commands.Runtime) int {
	name := filepath.Base(arguments[0])
	args := arguments[1:]
	if command, exists := lookup(name); exists && command.Linked {
		return dispatch(command, runtime, args)
	}
	if len(args) == 0 {
		usage(runtime)
		return commands.ExitUsage
	}
	if isHelpFlag(args[0]) || args[0] == "help" {
		return Help(runtime, args[1:])
	}
	command, exists := lookup(args[0])
	if !exists {
		fmt.Fprintf(runtime.Stderr, "dotfiles: commande inconnue « %s »\n", args[0])
		usage(runtime)
		return commands.ExitUsage
	}
	return dispatch(command, runtime, args[1:])
}

func dispatch(command Command, runtime commands.Runtime, args []string) int {
	if len(args) > 0 && isHelpFlag(args[len(args)-1]) {
		return Help(runtime, []string{command.Name})
	}
	return command.Run(runtime, args)
}

func runVerify(runtime commands.Runtime, args []string) int {
	root, code := repositoryRoot(args, runtime)
	if code != 0 {
		return code
	}
	return verify.Run(root, runtime.Stdout, runtime.Stderr)
}

func runHarnessAudit(runtime commands.Runtime, args []string) int {
	root, code := requiredRepositoryRoot(args, runtime)
	if code != 0 {
		return code
	}
	return audit.Run(auditConfig(root), runtime.Stdout, runtime.Stderr)
}

func runSkillRouting(runtime commands.Runtime, args []string) int {
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
	writeOverview(runtime.Stderr)
}
