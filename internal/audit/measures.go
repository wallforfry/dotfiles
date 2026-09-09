package audit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var descriptionBlock = regexp.MustCompile(`(?m)^description:.*\n(?:[ \t].*\n)*`)

func (a *auditor) measureContext() {
	a.section("Contexte toujours chargé", func() error {
		files := []string{"harness/AGENTS.md", "harness/SOUL.md", "harness/USER.md", "AGENTS.md", "dot_claude/CLAUDE.md"}
		total := 0
		for _, path := range files {
			size, err := fileSize(filepath.Join(a.config.Root, path))
			if err != nil {
				return fmt.Errorf("%s absent : coût du contexte non mesuré", path)
			}
			total += size
			fmt.Fprintf(a.stdout, "  %-34s %6d o  ~%5d jetons\n", path, size, size/4)
		}
		descriptions, err := descriptionBytes(a.config.Root)
		if err != nil {
			return err
		}
		total += descriptions
		fmt.Fprintf(a.stdout, "  %-34s %6d o  ~%5d jetons\n", "descriptions locales des skills", descriptions, descriptions/4)
		a.ok("%d octets, ~%d jetons sur chaque tâche de ce dépôt", total, total/4)
		fmt.Fprintln(a.stdout, "  Le CONTEXT.md de chaque hôte est chiffré : son coût n'est pas mesurable ici")
		fmt.Fprintln(a.stdout, "  Les descriptions de plugins gérées par l'hôte ne sont pas mesurables depuis le dépôt")
		return nil
	})
}

func fileSize(path string) (int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return int(info.Size()), nil
}

func descriptionBytes(root string) (int, error) {
	patterns := []string{"dot_config/agent-skills/*/SKILL.md", "dot_claude/agents/*.md"}
	total := 0
	for _, pattern := range patterns {
		files, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			return 0, err
		}
		for _, path := range files {
			content, err := os.ReadFile(path)
			if err != nil {
				return 0, fmt.Errorf("description illisible : %s", filepath.Base(path))
			}
			for _, match := range descriptionBlock.FindAll(content, -1) {
				total += len(match)
			}
		}
	}
	return total, nil
}

func (a *auditor) measureDeployment() {
	a.section("Source de déploiement", func() error {
		if _, err := exec.LookPath("chezmoi"); err != nil {
			return fmt.Errorf("chezmoi absent : retard de la source non mesuré")
		}
		source, err := commandOutput(a.config.Root, "chezmoi", "source-path")
		if err != nil {
			return fmt.Errorf("source chezmoi inconnue : retard non mesuré")
		}
		source = strings.TrimSpace(source)
		if samePath(source, a.config.Root) {
			a.ok("clone unique : la source de déploiement est ce dépôt")
			return nil
		}
		if _, err := commandOutput(source, "git", "rev-parse", "--verify", "-q", "origin/main"); err != nil {
			return fmt.Errorf("source de déploiement : origin/main inconnue, retard non mesuré")
		}
		behind, err := commandOutput(source, "git", "rev-list", "--count", "HEAD..origin/main")
		if err != nil {
			return fmt.Errorf("source de déploiement : calcul du retard interrompu")
		}
		a.ok("source en retard de %s commits sur origin/main à la dernière récupération", strings.TrimSpace(behind))
		return nil
	})
}

func gitRepository(root string) bool {
	_, err := commandOutput(root, "git", "rev-parse", "--show-toplevel")
	return err == nil
}

func commandOutput(directory, name string, arguments ...string) (string, error) {
	command := exec.Command(name, arguments...)
	command.Dir = directory
	output, err := command.Output()
	return string(output), err
}

func samePath(left, right string) bool {
	leftPath, leftErr := filepath.EvalSymlinks(left)
	rightPath, rightErr := filepath.EvalSymlinks(right)
	return leftErr == nil && rightErr == nil && filepath.Clean(leftPath) == filepath.Clean(rightPath)
}
