package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

type frontmatter struct {
	raw         string
	name        string
	description string
	category    string
}

func parseFrontmatter(content string) frontmatter {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return frontmatter{}
	}
	var block []string
	for _, line := range lines[1:] {
		if line == "---" {
			break
		}
		block = append(block, line)
	}
	fm := frontmatter{raw: strings.Join(block, "\n")}
	for index, line := range block {
		switch {
		case strings.HasPrefix(line, "name:"):
			fm.name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		case strings.HasPrefix(line, "description:"):
			first := strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			first = strings.TrimSpace(strings.TrimLeft(first, ">-"))
			parts := []string{}
			if first != "" {
				parts = append(parts, first)
			}
			for _, continuation := range block[index+1:] {
				if continuation != "" && continuation[0] != ' ' && continuation[0] != '\t' {
					break
				}
				if value := strings.TrimSpace(continuation); value != "" {
					parts = append(parts, value)
				}
			}
			fm.description = strings.Join(parts, " ")
		case strings.TrimSpace(line) != line && strings.HasPrefix(strings.TrimSpace(line), "category:"):
			fm.category = strings.Trim(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "category:")), `"'`)
		}
	}
	return fm
}

type readmeSkill struct {
	category    string
	description string
}

func parseReadmeSkills(content string) map[string]readmeSkill {
	result := map[string]readmeSkill{}
	category := ""
	linePattern := regexp.MustCompile(`^\| \x60([a-z-]+)\x60 \| (.*) \|$`)
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "## ") {
			category = strings.ToLower(strings.Fields(line)[1])
			continue
		}
		match := linePattern.FindStringSubmatch(line)
		if match != nil {
			result[match[1]] = readmeSkill{category: category, description: match[2]}
		}
	}
	return result
}

func derivedDescription(value string) string {
	if index := strings.Index(value, ". "); index >= 0 {
		return value[:index+1]
	}
	return value + "."
}

func (v *verifier) checkSkills() {
	v.head("Skills")
	readme, err := os.ReadFile(v.path("dot_config", "agent-skills", "README.md"))
	if err != nil {
		v.ko("dot_config/agent-skills/README.md absent ou illisible")
		return
	}
	listed := parseReadmeSkills(string(readme))
	dirs := globDirs(v.path("dot_config", "agent-skills", "*"))
	for _, dir := range dirs {
		slug := filepath.Base(filepath.Clean(dir))
		fm, exists := v.validateSkillDocument(dir, slug)
		if !exists {
			continue
		}
		v.validateSkillIndex(slug, fm, listed)
		v.validateSkillLayout(dir, slug)
	}
	if len(listed) != len(dirs) {
		v.ko(fmt.Sprintf("le tableau du README annonce %d skills pour %d répertoires", len(listed), len(dirs)))
	}
	for slug := range listed {
		if info, err := os.Stat(v.path("dot_config", "agent-skills", slug)); err != nil || !info.IsDir() {
			v.ko(fmt.Sprintf("%s : ligne du README sans répertoire", slug))
		}
	}
	v.checkMergeVerdictEvidenceStatus()
	v.okIf(fmt.Sprintf("%d skills, frontmatter et tableau cohérents", len(dirs)))
}

func (v *verifier) checkMergeVerdictEvidenceStatus() {
	path := v.path("dot_config", "agent-skills", "merge-verdict", "references", "cases.md")
	content, err := os.ReadFile(path)
	if err != nil {
		v.ko("merge-verdict : cas de référence absent ou illisible")
		return
	}
	line := "**Evidence status:** observation only, not reproducible evidence for either approval verdict."
	if !slices.Contains(strings.Split(string(content), "\n"), line) {
		v.ko("merge-verdict : les observations C et D peuvent être prises pour une preuve d'approbation")
	}
}

func (v *verifier) validateSkillDocument(dir, slug string) (frontmatter, bool) {
	content, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		v.ko(fmt.Sprintf("%s : SKILL.md manquant", slug))
		return frontmatter{}, false
	}
	fm := parseFrontmatter(string(content))
	if fm.raw == "" {
		v.ko(fmt.Sprintf("%s : frontmatter absent", slug))
	}
	if fm.name != slug {
		v.ko(fmt.Sprintf("%s : frontmatter name=« %s » ne correspond pas au répertoire", slug, fm.name))
	}
	if !strings.Contains(fm.raw, "description:") {
		v.ko(fmt.Sprintf("%s : pas de description", slug))
	}
	if !strings.Contains(string(content), "Use when") {
		v.ko(fmt.Sprintf("%s : la description ne dit pas quand l'utiliser", slug))
	}
	if fm.category != "dev" && fm.category != "ops" {
		category := fm.category
		if category == "" {
			category = "absent"
		}
		v.ko(fmt.Sprintf("%s : metadata.category=« %s » hors de {dev, ops}", slug, category))
	}
	return fm, true
}

func (v *verifier) validateSkillIndex(slug string, fm frontmatter, listed map[string]readmeSkill) {
	row, exists := listed[slug]
	if !exists {
		v.ko(fmt.Sprintf("%s : absent du tableau de dot_config/agent-skills/README.md", slug))
		return
	}
	if row.category != fm.category {
		v.ko(fmt.Sprintf("%s : listé sous « %s » dans le README pour une catégorie « %s »", slug, row.category, fm.category))
	}
	expected := derivedDescription(fm.description)
	if row.description == "" {
		v.ko(fmt.Sprintf("%s : ligne du README illisible, dérivation non vérifiée", slug))
	} else if row.description != expected {
		v.ko(fmt.Sprintf("%s : la ligne du README ne dérive plus de la description du frontmatter", slug))
		fmt.Fprintf(v.stderr, "      README : %s\n      SKILL  : %s\n", row.description, expected)
	}
}

func (v *verifier) validateSkillLayout(dir, slug string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		v.ko(fmt.Sprintf("%s : répertoire illisible", slug))
		return
	}
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "references" && entry.Name() != "assets" && entry.Name() != "scripts" {
			v.ko(fmt.Sprintf("%s : sous-répertoire « %s » hors de la liste", slug, entry.Name()))
		}
	}
}
