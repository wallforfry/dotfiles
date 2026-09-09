package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func (v *verifier) checkSubagents() {
	v.head("Subagents")
	n := 0
	for _, file := range glob(v.path("dot_claude", "agents", "*.md")) {
		content, err := os.ReadFile(file)
		if err != nil {
			v.ko(fmt.Sprintf("%s illisible", relative(v.root, file)))
			continue
		}
		slug := strings.TrimSuffix(filepath.Base(file), ".md")
		fm := parseFrontmatter(string(content))
		if fm.name != slug {
			v.ko(fmt.Sprintf("%s : frontmatter name=« %s » ne correspond pas au fichier", slug, fm.name))
		}
		if !strings.Contains(fm.raw, "description:") {
			v.ko(fmt.Sprintf("%s : pas de description", slug))
		}
		if !strings.Contains(fm.raw, "Use when") {
			v.ko(fmt.Sprintf("%s : « Use when » absent de la description", slug))
		}
		if !strings.Contains(fm.raw, "tools:") {
			v.ko(fmt.Sprintf("%s : pas de champ tools", slug))
		}
		n++
	}
	if info, err := os.Stat(v.path("dot_claude", "agents")); err == nil && info.IsDir() {
		if _, err := os.Lstat(v.path("dot_claude_pro", "symlink_agents")); err != nil {
			v.ko("dot_claude/agents existe sans symlink_agents dans dot_claude_pro : le profil pro ne les verrait pas")
		}
	}
	v.okIf(fmt.Sprintf("%d subagents, frontmatter cohérent et partagé avec le profil pro", n))
}

func (v *verifier) checkWorkflows() {
	v.head("Workflows")
	files, err := v.gitLines("ls-files", ".github/*")
	if err != nil {
		v.ko("inventaire des workflows interrompu")
		return
	}
	verbose := regexp.MustCompile(`^[^#]*(--verbose|chezmoi[[:space:]]+diff)`)
	cat := regexp.MustCompile(`^[^#]*chezmoi[^#]*[[:space:]]cat[[:space:]]`)
	redirected := regexp.MustCompile(`>[[:space:]]*("?\$|/dev/null)`)
	for _, file := range files {
		content, err := os.ReadFile(v.path(file))
		if err != nil {
			v.ko(fmt.Sprintf("%s illisible", file))
			continue
		}
		var hits []string
		for index, line := range strings.Split(string(content), "\n") {
			if verbose.MatchString(line) || (cat.MatchString(line) && !redirected.MatchString(line)) {
				hits = append(hits, fmt.Sprintf("%d:%s", index+1, line))
			}
		}
		if len(hits) > 0 {
			v.ko(fmt.Sprintf("%s : commande qui écrirait le contenu rendu d'une cible dans un log public", file))
			fmt.Fprint(v.stderr, indent(firstLines(strings.Join(hits, "\n"), 3)))
		}
	}
	v.okIf(fmt.Sprintf("%d fichiers de CI, aucune sortie de contenu rendu", len(files)))
}

func (v *verifier) checkADR() {
	v.head("Index des ADR")
	entries, err := os.ReadDir(v.path("docs", "adr"))
	if err != nil {
		v.ko("répertoire des ADR illisible")
		return
	}
	pattern := regexp.MustCompile(`^[0-9]{3}`)
	var actual []string
	for _, entry := range entries {
		if prefix := pattern.FindString(entry.Name()); prefix != "" {
			actual = append(actual, prefix)
		}
	}
	readme, err := os.ReadFile(v.path("docs", "adr", "README.md"))
	if err != nil {
		v.ko("index des ADR illisible")
		return
	}
	indexPattern := regexp.MustCompile(`(?m)^\| \[([0-9]{3})\]`)
	var indexed []string
	for _, match := range indexPattern.FindAllStringSubmatch(string(readme), -1) {
		indexed = append(indexed, match[1])
	}
	sort.Strings(actual)
	if strings.Join(actual, "\n") != strings.Join(indexed, "\n") {
		v.ko("l'index des ADR ne coïncide pas avec le répertoire")
		return
	}
	v.ok(fmt.Sprintf("%d ADR, index et répertoire coïncident", len(actual)))
}

func (v *verifier) checkEncryption() {
	v.head("Fragments chiffrés")
	files, err := v.gitLines("ls-files", "*.age")
	if err != nil {
		v.ko("inventaire des fragments chiffrés interrompu")
		return
	}
	n := 0
	for _, file := range files {
		content, err := os.ReadFile(v.path(file))
		if err != nil {
			v.ko(fmt.Sprintf("%s illisible", file))
			continue
		}
		prefix := content
		if len(prefix) > 40 {
			prefix = prefix[:40]
		}
		if strings.Contains(string(prefix), "BEGIN AGE ENCRYPTED") || strings.Contains(string(prefix), "age-encryption.org") {
			n++
		} else {
			v.ko(fmt.Sprintf("%s n'est pas chiffré", file))
		}
	}
	v.okIf(fmt.Sprintf("%d fragments, tous chiffrés", n))
}

func (v *verifier) checkLiveState() {
	v.head("État vivant préservé")
	n := 0
	for _, dir := range []string{"dot_claude", "dot_claude_pro", "dot_codex", "private_dot_gnupg"} {
		path := v.path(dir)
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			continue
		}
		n++
		found := false
		_ = filepath.WalkDir(path, func(_ string, entry os.DirEntry, err error) error {
			if err == nil && strings.HasPrefix(entry.Name(), "exact_") {
				found = true
				return filepath.SkipAll
			}
			return nil
		})
		if found {
			v.ko(fmt.Sprintf("%s : attribut exact_ présent, chezmoi supprimerait l'état vivant", dir))
		}
	}
	for _, dir := range []string{"exact_dot_claude", "exact_dot_claude_pro", "exact_dot_codex", "exact_private_dot_gnupg"} {
		if _, err := os.Lstat(v.path(dir)); err == nil {
			v.ko(fmt.Sprintf("%s : racine préfixée exact_, chezmoi supprimerait l'état vivant", dir))
		}
	}
	v.okIf(fmt.Sprintf("%d racines d'état vivant, aucun attribut exact_", n))
}
