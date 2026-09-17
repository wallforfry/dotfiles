package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (v *verifier) render(config, source string) ([]byte, error) {
	content, err := os.ReadFile(source)
	if err != nil {
		return nil, err
	}
	return v.command(content, "chezmoi", "execute-template", "--config", config, "--source", v.root)
}

func (v *verifier) checkSyntax() {
	v.head("Syntaxe des scripts")
	n := 0
	for _, file := range glob(v.path("run_*.sh.tmpl")) {
		valid := true
		for _, config := range v.configs {
			rendered, err := v.render(config, file)
			if err != nil {
				v.ko(fmt.Sprintf("%s ne se rend pas avec %s", filepath.Base(file), filepath.Base(config)))
				valid = false
				break
			}
			if output, err := v.command(rendered, "sh", "-n"); err != nil {
				v.ko(fmt.Sprintf("%s ne passe pas sh -n avec %s\n%s", filepath.Base(file), filepath.Base(config), indent(firstLines(string(output), 3))))
				valid = false
				break
			}
		}
		if valid {
			n++
		}
	}
	patterns := []string{"scripts/*.sh", "dot_claude/hooks/executable_*", "dot_local/bin/executable_*"}
	for _, pattern := range patterns {
		for _, file := range glob(v.path(pattern)) {
			content, err := os.ReadFile(file)
			if err != nil {
				v.ko(fmt.Sprintf("%s illisible", relative(v.root, file)))
				continue
			}
			checker := "sh"
			if first := strings.SplitN(string(content), "\n", 2)[0]; strings.Contains(first, "bash") {
				checker = "bash"
			}
			if output, err := v.command(nil, checker, "-n", file); err != nil {
				v.ko(fmt.Sprintf("%s ne passe pas %s -n\n%s", relative(v.root, file), checker, indent(firstLines(string(output), 3))))
				continue
			}
			n++
		}
	}
	v.checkScriptBoundary()
	v.okIf(fmt.Sprintf("%d scripts, syntaxe valide sur %d combinaisons de profil", n, len(v.configs)))
}

func (v *verifier) checkScriptBoundary() {
	files, err := v.gitLines("ls-files", "--cached", "--others", "--exclude-standard")
	if err != nil {
		v.ko("inventaire des scripts interrompu")
		return
	}
	allowed := map[string]bool{
		"run_before_unlock-age-key.sh.tmpl":         true,
		"run_onchange_before_install-tools.sh.tmpl": true,
		"run_onchange_after_build-dotfiles.sh.tmpl": true,
	}
	for _, file := range files {
		if allowed[file] {
			continue
		}
		if strings.HasSuffix(file, ".py") || strings.HasSuffix(file, ".sh") || strings.HasSuffix(file, ".sh.tmpl") {
			v.ko(fmt.Sprintf("%s sort de la frontière de bootstrap Go", file))
			continue
		}
		content, readErr := os.ReadFile(v.path(file))
		if readErr == nil && scriptShebang(content) {
			v.ko(fmt.Sprintf("%s contient un script hors bootstrap", file))
		}
	}
}

func scriptShebang(content []byte) bool {
	first := strings.ToLower(strings.SplitN(string(content), "\n", 2)[0])
	return strings.HasPrefix(first, "#!") &&
		(strings.Contains(first, "python") || strings.Contains(first, "/sh") || strings.Contains(first, "bash"))
}

func (v *verifier) checkTemplates() {
	v.head("Rendu des templates")
	files, err := v.gitLines("ls-files", "--cached", "--others", "--exclude-standard", "--", "*.tmpl")
	if err != nil {
		v.ko("inventaire des templates interrompu")
		return
	}
	n := 0
	for _, file := range files {
		if file == ".chezmoi.toml.tmpl" {
			continue
		}
		if _, err := os.Stat(v.path(file)); os.IsNotExist(err) {
			continue
		}
		valid := true
		for _, config := range v.configs {
			if _, err := v.render(config, v.path(file)); err != nil {
				v.ko(fmt.Sprintf("%s ne se rend pas avec %s", file, filepath.Base(config)))
				valid = false
				break
			}
		}
		if valid {
			n++
		}
	}
	v.okIf(fmt.Sprintf("%d templates rendus sur %d combinaisons", n, len(v.configs)))
}

func relative(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}
