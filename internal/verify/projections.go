package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func containsLine(path string, needle string) bool {
	content, err := os.ReadFile(path)
	return err == nil && strings.Contains(string(content), needle)
}

func (v *verifier) checkProjections() {
	v.head("Projections d'instructions")
	n := v.checkCanonicalSources()
	v.checkProjectionIncludes()
	v.checkSkillLinks()
	v.checkContextProjection()
	v.okIf(fmt.Sprintf("%d sources canoniques sans import, projections et liens de skills cohérents", n))
}

func (v *verifier) checkCanonicalSources() int {
	n := 0
	importPattern := regexp.MustCompile(`(?m)^@[A-Za-z]`)
	for _, file := range glob(v.path("harness", "*.md")) {
		content, err := os.ReadFile(file)
		if err != nil {
			v.ko(fmt.Sprintf("%s illisible", relative(v.root, file)))
			continue
		}
		if importPattern.Match(content) {
			v.ko(fmt.Sprintf("%s : import « @ » propre à un seul hôte dans une source agnostique", relative(v.root, file)))
			var hits []string
			for index, line := range strings.Split(string(content), "\n") {
				if importPattern.MatchString(line) {
					hits = append(hits, fmt.Sprintf("%d:%s", index+1, line))
				}
			}
			fmt.Fprint(v.stderr, indent(firstLines(strings.Join(hits, "\n"), 3)))
		}
		n++
	}
	return n
}

func (v *verifier) checkProjectionIncludes() {
	for _, base := range []string{"AGENTS", "SOUL", "USER"} {
		projection := v.path("dot_claude", base+".md.tmpl")
		content, err := os.ReadFile(projection)
		if err != nil {
			v.ko(fmt.Sprintf("dot_claude/%s.md.tmpl absent : projection Claude manquante", base))
			continue
		}
		if len(strings.Split(strings.TrimSuffix(string(content), "\n"), "\n")) != 1 {
			v.ko(fmt.Sprintf("dot_claude/%s.md.tmpl : une projection ne porte qu'une ligne d'include", base))
		}
		include := fmt.Sprintf(`include "harness/%s.md"`, base)
		if !strings.Contains(string(content), include) {
			v.ko(fmt.Sprintf("dot_claude/%s.md.tmpl : n'inclut pas harness/%s.md", base, base))
		}
		if !containsLine(v.path("dot_claude", "CLAUDE.md"), "@"+base+".md") {
			v.ko(fmt.Sprintf("dot_claude/CLAUDE.md : @%s.md absent, Claude ne chargerait pas %s", base, base))
		}
		if !containsLine(v.path("dot_codex", "AGENTS.md.tmpl"), include) {
			v.ko(fmt.Sprintf("dot_codex/AGENTS.md.tmpl : n'inclut pas harness/%s.md", base))
		}
	}
}

func (v *verifier) checkSkillLinks() {
	for _, host := range []string{"dot_claude", "dot_codex"} {
		if _, err := os.Lstat(v.path(host, "symlink_skills")); err == nil {
			v.ko(fmt.Sprintf("%s/symlink_skills : un lien sur le répertoire entier détruirait l'état vivant", host))
		}
	}
	for _, dir := range globDirs(v.path("dot_config", "agent-skills", "*")) {
		slug := filepath.Base(filepath.Clean(dir))
		for _, host := range []string{"dot_claude", "dot_codex"} {
			link := v.path(host, "skills", "symlink_"+slug)
			content, err := os.ReadFile(link)
			if err != nil {
				v.ko(fmt.Sprintf("%s absent : %s ne verrait pas la skill %s", relative(v.root, link), host, slug))
			} else if strings.TrimSuffix(string(content), "\n") != "../../.config/agent-skills/"+slug {
				v.ko(fmt.Sprintf("%s : cible hors de la source unique de skills", relative(v.root, link)))
			}
		}
	}
	for _, host := range []string{"dot_claude", "dot_codex"} {
		for _, link := range glob(v.path(host, "skills", "symlink_*")) {
			slug := strings.TrimPrefix(filepath.Base(link), "symlink_")
			if info, err := os.Stat(v.path("dot_config", "agent-skills", slug)); err != nil || !info.IsDir() {
				v.ko(fmt.Sprintf("%s : lien sans skill correspondante", relative(v.root, link)))
			}
		}
	}
	professional, err := os.ReadFile(v.path("dot_claude_pro", "symlink_skills"))
	if err != nil || strings.TrimSuffix(string(professional), "\n") != "../.claude/skills" {
		v.ko("dot_claude_pro/symlink_skills : cible hors de ~/.claude/skills")
	}
}

func (v *verifier) checkContextProjection() {
	if containsLine(v.path("dot_codex", "AGENTS.md.tmpl"), "CONTEXT.md") {
		if _, err := os.Stat(v.path("dot_codex", "symlink_CONTEXT.md")); err != nil {
			v.ko("dot_codex : l'adaptateur renvoie vers CONTEXT.md sans symlink_CONTEXT.md")
		}
	}
}
