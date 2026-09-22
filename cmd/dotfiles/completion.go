package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/wallforfry/dotfiles/internal/commands"
)

// Completion écrit sur la sortie standard le script de complétion du shell
// demandé, engendré depuis le catalogue.
func Completion(runtime commands.Runtime, args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(runtime.Stderr, "dotfiles: usage - completion zsh|bash")
		return commands.ExitUsage
	}
	switch args[0] {
	case "zsh":
		writeZshCompletion(runtime.Stdout)
	case "bash":
		writeBashCompletion(runtime.Stdout)
	default:
		fmt.Fprintf(runtime.Stderr, "dotfiles: shell non pris en charge « %s »\n", args[0])
		return commands.ExitUsage
	}
	return 0
}

func linkedCommands() []Command {
	linked := make([]Command, 0, len(catalog))
	for _, command := range catalog {
		if command.Linked && len(command.Subcommands) > 0 {
			linked = append(linked, command)
		}
	}
	return linked
}

func shellFunctionName(prefix, name string) string {
	return prefix + strings.ReplaceAll(name, "-", "_")
}

func writeZshCompletion(writer io.Writer) {
	fmt.Fprint(writer, "_dotfiles() {\n  local -a _dotfiles_commands\n  _dotfiles_commands=(\n")
	for _, command := range catalog {
		fmt.Fprintf(writer, "    %s\n", shellQuote(command.Name+":"+command.Summary))
	}
	fmt.Fprint(writer, "  )\n  if (( CURRENT == 2 )); then\n    _describe 'commande' _dotfiles_commands\n    return\n  fi\n  case ${words[2]} in\n")
	for _, command := range catalog {
		if completion := zshArguments(command); completion != "" {
			fmt.Fprintf(writer, "    %s) %s ;;\n", command.Name, completion)
		}
	}
	fmt.Fprint(writer, "    *) ;;\n  esac\n}\ncompdef _dotfiles dotfiles\n")
	for _, command := range linkedCommands() {
		function := shellFunctionName("_dotfiles_", command.Name)
		fmt.Fprintf(writer, "\n%s() {\n  %s\n}\ncompdef %s %s\n", function, zshArguments(command), function, command.Name)
	}
}

func writeBashCompletion(writer io.Writer) {
	names := make([]string, 0, len(catalog))
	for _, command := range catalog {
		names = append(names, command.Name)
	}
	fmt.Fprintf(writer, "_dotfiles() {\n  local current=${COMP_WORDS[COMP_CWORD]}\n  if [ \"$COMP_CWORD\" -eq 1 ]; then\n    COMPREPLY=( $(compgen -W %s -- \"$current\") )\n    return\n  fi\n  case ${COMP_WORDS[1]} in\n", shellQuote(strings.Join(names, " ")))
	for _, command := range catalog {
		if completion := bashArguments(command); completion != "" {
			fmt.Fprintf(writer, "    %s) %s ;;\n", command.Name, completion)
		}
	}
	fmt.Fprint(writer, "    *) COMPREPLY=() ;;\n  esac\n}\ncomplete -F _dotfiles dotfiles\n")
	for _, command := range linkedCommands() {
		function := shellFunctionName("_dotfiles_", command.Name)
		fmt.Fprintf(writer, "\n%s() {\n  local current=${COMP_WORDS[COMP_CWORD]}\n  %s\n}\ncomplete -F %s %s\n", function, bashArguments(command), function, command.Name)
	}
}

// bashArguments et zshArguments rendent vide pour une commande sans argument :
// mieux vaut ne rien proposer qu'un chemin quelconque.
func bashArguments(command Command) string {
	if len(command.Subcommands) == 0 && !command.Paths {
		return ""
	}
	files := ""
	if command.Paths {
		files = "-f "
	}
	return fmt.Sprintf("COMPREPLY=( $(compgen -W %s %s-- \"$current\") )", shellQuote(strings.Join(command.Subcommands, " ")), files)
}

func zshArguments(command Command) string {
	words := ""
	if len(command.Subcommands) > 0 {
		words = "compadd -- " + zshWords(command.Subcommands)
	}
	if !command.Paths {
		return words
	}
	if words == "" {
		return "_files"
	}
	return words + "; _files"
}

func zshWords(words []string) string {
	quoted := make([]string, 0, len(words))
	for _, word := range words {
		quoted = append(quoted, shellQuote(word))
	}
	return strings.Join(quoted, " ")
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}
