package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/wallforfry/dotfiles/internal/commands"
)

func isHelpFlag(argument string) bool {
	return argument == "-h" || argument == "--help"
}

// Help écrit l'aide générale, ou celle d'une commande nommée.
func Help(runtime commands.Runtime, args []string) int {
	if len(args) == 0 {
		writeOverview(runtime.Stdout)
		return 0
	}
	if len(args) > 1 {
		fmt.Fprintln(runtime.Stderr, "dotfiles: usage - help [commande]")
		return commands.ExitUsage
	}
	command, exists := lookup(args[0])
	if !exists {
		fmt.Fprintf(runtime.Stderr, "dotfiles: commande inconnue « %s »\n", args[0])
		writeOverview(runtime.Stderr)
		return commands.ExitUsage
	}
	fmt.Fprintf(runtime.Stdout, "%s - %s\n\n%s\n", command.Name, command.Summary, command.Usage)
	return 0
}

func writeOverview(writer io.Writer) {
	fmt.Fprint(writer, "dotfiles - outillage du dépôt de configuration\n\nusage : dotfiles <commande> [arguments]\n\ncommandes :\n")
	width := 0
	for _, command := range catalog {
		if length := len(command.Name); length > width {
			width = length
		}
	}
	for _, command := range catalog {
		fmt.Fprintf(writer, "  %s%s  %s\n", command.Name, strings.Repeat(" ", width-len(command.Name)), command.Summary)
	}
	fmt.Fprint(writer, "\n« dotfiles help <commande> » détaille une commande.\n")
}
