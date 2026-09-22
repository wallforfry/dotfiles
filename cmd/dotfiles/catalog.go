package main

import "github.com/wallforfry/dotfiles/internal/commands"

// Command décrit une commande du CLI. Le catalogue est à la fois la table de
// routage et la source dont dérivent l'aide, la ligne d'usage et les scripts de
// complétion : une commande sans entrée ici n'est pas routée du tout.
type Command struct {
	Name string
	// Run exécute la commande. Il est posé par newCatalog, hors de la valeur
	// littérale, parce que deux commandes ont besoin du catalogue lui-même.
	Run         func(commands.Runtime, []string) int
	Summary     string
	Usage       string
	Subcommands []string
	// Linked vaut vrai quand ~/.local/bin porte un lien du même nom vers le CLI.
	Linked bool
	// Paths vaut vrai quand un argument de la commande est un chemin.
	Paths bool
}

func commandList() []Command {
	return []Command{
		{
			Name:    "help",
			Summary: "affiche cette aide, ou le détail d'une commande",
			Usage:   "dotfiles help [commande]",
		},
		{
			Name:    "completion",
			Summary: "écrit un script de complétion sur la sortie standard",
			Usage: "dotfiles completion zsh|bash\n\n" +
				"Le script est engendré depuis le catalogue des commandes.\n" +
				"zsh : eval \"$(dotfiles completion zsh)\"\n" +
				"bash : eval \"$(dotfiles completion bash)\"",
			Subcommands: []string{"zsh", "bash"},
		},
		{
			Name:    "verify",
			Paths:   true,
			Summary: "rejoue la barrière mécanique du dépôt",
			Usage: "dotfiles verify [--repository <chemin>]\n\n" +
				"Sans --repository, la racine est celle du dépôt git courant.",
			Subcommands: []string{"--repository"},
		},
		{
			Name:    "harness-audit",
			Paths:   true,
			Summary: "mesure le coût et l'activation du harness",
			Usage: "dotfiles harness-audit --repository <chemin>\n\n" +
				"Variables lues : CLAUDE_PROJECTS, CODEX_SESSIONS, HARNESS_RULES_SINCE.",
			Subcommands: []string{"--repository"},
		},
		{
			Name:    "validate-skill-routing",
			Paths:   true,
			Summary: "vérifie le routage des skills sur un corpus",
			Usage:   "dotfiles validate-skill-routing [corpus.tsv]",
		},
		{
			Name:    "register-claude-hook",
			Summary: "enregistre un hook dans la configuration Claude Code",
			Usage:   "dotfiles register-claude-hook",
		},
		{
			Name:        "register-hindsight",
			Paths:       true,
			Summary:     "enregistre la configuration Hindsight du coding-agent",
			Usage:       "dotfiles register-hindsight --config <fichier JSON déployé>",
			Subcommands: []string{"--config"},
		},
		{
			Name:        "hindsight",
			Summary:     "gère les banques de mémoire Hindsight",
			Paths:       true,
			Usage:       "dotfiles hindsight bank create <banque>\ndotfiles hindsight bank add <dossier> <banque>\ndotfiles hindsight bank remove <dossier>",
			Subcommands: []string{"bank"},
		},
		{
			Name:    "agent-handoff",
			Summary: "prépare la passation d'une session d'agent",
			Usage:   "dotfiles agent-handoff",
		},
		{
			Name:        "cloak",
			Summary:     "pilote le conteneur du navigateur exposé aux MCP",
			Usage:       "cloak --start | --stop | --status | --url",
			Subcommands: []string{"--start", "--stop", "--status", "--url"},
			Linked:      true,
		},
		{
			Name:    "firecrawl-mcp",
			Summary: "lance le serveur MCP Firecrawl conteneurisé",
			Usage:   "firecrawl-mcp",
			Linked:  true,
		},
		{
			Name:    "postgres-mcp",
			Summary: "lance le serveur MCP PostgreSQL conteneurisé",
			Usage:   "postgres-mcp",
			Linked:  true,
		},
		{
			Name:    "scrapling-mcp",
			Summary: "lance le serveur MCP Scrapling conteneurisé",
			Usage:   "scrapling-mcp",
			Linked:  true,
		},
		{
			Name:        "smartcard-wakeup",
			Summary:     "réveille la carte à puce GPG",
			Usage:       "smartcard-wakeup [--hook]",
			Subcommands: []string{"--hook"},
			Linked:      true,
		},
	}
}

// runner relie un nom à son exécution. La table est distincte du littéral parce
// que l'aide et la complétion lisent le catalogue, ce qui interdirait un cycle
// d'initialisation ; un test exige que les deux tables se recouvrent.
func runner(name string) (func(commands.Runtime, []string) int, bool) {
	found, exists := runners()[name]
	return found, exists
}

func runners() map[string]func(commands.Runtime, []string) int {
	return map[string]func(commands.Runtime, []string) int{
		"help":                   Help,
		"completion":             Completion,
		"verify":                 runVerify,
		"harness-audit":          runHarnessAudit,
		"validate-skill-routing": runSkillRouting,
		"register-claude-hook":   commands.RegisterClaudeHook,
		"register-hindsight":     commands.RegisterHindsight,
		"hindsight":              commands.HindsightBank,
		"agent-handoff":          commands.AgentHandoff,
		"cloak":                  commands.Cloak,
		"firecrawl-mcp":          commands.Firecrawl,
		"postgres-mcp":           commands.Postgres,
		"scrapling-mcp":          commands.Scrapling,
		"smartcard-wakeup":       commands.SmartcardWakeup,
	}
}

func lookup(name string) (Command, bool) {
	for _, command := range commandList() {
		if command.Name != name {
			continue
		}
		found, exists := runner(name)
		if !exists || found == nil {
			return Command{}, false
		}
		command.Run = found
		return command, true
	}
	return Command{}, false
}
