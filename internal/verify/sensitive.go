package verify

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var ipPattern = regexp.MustCompile(`(^|[^0-9.])((1[0-9]{2}|2[0-9]{2}|[1-9][0-9]?)\.){3}(1[0-9]{2}|2[0-9]{2}|[1-9][0-9]?)`)

func commandExitCode(err error) int {
	if err == nil {
		return 0
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode()
	}
	return -1
}

func loadSensitivePatterns(path string) (string, *regexp.Regexp, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", nil, fmt.Errorf("%s absent ou illisible : le contrôle des noms sensibles n'a pas été fait", path)
	}
	var patterns []string
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			patterns = append(patterns, line)
		}
	}
	if len(patterns) == 0 {
		return "", nil, fmt.Errorf("%s vide : le contrôle des noms sensibles n'a pas été fait", path)
	}
	joined := strings.Join(patterns, "|")
	compiled, err := regexp.Compile("(?i)" + joined)
	if err != nil {
		return "", nil, fmt.Errorf("%s contient une expression invalide : %w", path, err)
	}
	return joined, compiled, nil
}

func disallowedIPLines(content string) []string {
	var hits []string
	for _, line := range nonemptyLines(content) {
		if ipPattern.MatchString(line) && !strings.Contains(line, "127.0.0.1") && !strings.Contains(line, "0.0.0.0") {
			hits = append(hits, line)
		}
	}
	return hits
}

func (v *verifier) checkSensitive() {
	v.head("Rien de sensible en clair (ADR-016)")
	list, err := sensitiveListPath()
	if err != nil {
		v.ko(err.Error())
		return
	}
	patterns, compiled, err := loadSensitivePatterns(list)
	if err != nil {
		v.ko(err.Error())
	}
	v.checkTrackedSensitive(patterns, compiled != nil)
	if compiled == nil {
		return
	}
	v.checkUntrackedSensitive(compiled)
	v.checkSensitivePaths(compiled)
	v.checkSensitiveBranch(compiled)
	v.checkSensitiveCommits(compiled)
}

func sensitiveListPath() (string, error) {
	if list := os.Getenv("SENSIBLE_LIST"); list != "" {
		return list, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("répertoire personnel introuvable : le contrôle des noms sensibles n'a pas été fait")
	}
	return filepath.Join(home, ".config", "dotfiles", "sensible.txt"), nil
}

func (v *verifier) checkTrackedSensitive(patterns string, enabled bool) {
	trackedHits := 0
	if enabled {
		output, grepErr := v.command(nil, "git", "grep", "-IlE", patterns, "--", ".", ":!docs/adr/016-*")
		switch code := commandExitCode(grepErr); {
		case code > 1 || code < 0:
			v.ko("arbre : lecture interrompue")
		case code == 0:
			trackedHits += len(nonemptyLines(string(output)))
		}
		output, grepErr = v.command(nil, "git", "grep", "-InE", ipPattern.String(), "--", ".")
		switch code := commandExitCode(grepErr); {
		case code > 1 || code < 0:
			v.ko("arbre : lecture des adresses interrompue")
		case code == 0:
			trackedHits += len(disallowedIPLines(string(output)))
		}
	}
	if trackedHits > 0 {
		v.ko(fmt.Sprintf("arbre : %d occurrence(s)", trackedHits))
	} else {
		v.okIf("arbre : aucune occurrence")
	}
}

func (v *verifier) checkUntrackedSensitive(compiled *regexp.Regexp) {
	untracked, err := v.gitNullFiles("ls-files", "-z", "--others", "--exclude-standard")
	if err != nil {
		v.ko("contenu des fichiers non suivis : inventaire interrompu")
	} else {
		hits, unreadable := 0, 0
		for _, file := range untracked {
			content, readErr := os.ReadFile(v.path(file))
			if readErr != nil {
				unreadable++
				continue
			}
			if bytes.IndexByte(content, 0) >= 0 {
				continue
			}
			text := string(content)
			if compiled.MatchString(text) || len(disallowedIPLines(text)) > 0 {
				hits++
			}
		}
		if unreadable > 0 {
			v.ko(fmt.Sprintf("contenu des fichiers non suivis : %d lecture(s) impossible(s)", unreadable))
		} else if hits > 0 {
			v.ko(fmt.Sprintf("contenu des fichiers non suivis : %d occurrence(s)", hits))
		} else {
			v.okIf("contenu des fichiers non suivis : aucune occurrence")
		}
	}
}

func (v *verifier) checkSensitivePaths(compiled *regexp.Regexp) {
	paths, err := v.gitLines("ls-files", "--cached", "--others", "--exclude-standard")
	if err != nil {
		v.ko("noms de fichiers : inventaire interrompu")
	} else {
		pathHits := 0
		for _, path := range paths {
			if compiled.MatchString(path) {
				pathHits++
			}
		}
		if pathHits > 0 {
			v.ko(fmt.Sprintf("noms de fichiers suivis ou non ignorés : %d occurrence(s)", pathHits))
		} else {
			v.okIf("noms de fichiers suivis ou non ignorés : aucune occurrence")
		}
	}
}

func (v *verifier) checkSensitiveBranch(compiled *regexp.Regexp) {
	branchOutput, branchErr := v.command(nil, "git", "symbolic-ref", "--quiet", "--short", "HEAD")
	branch := strings.TrimSpace(string(branchOutput))
	if branchErr != nil {
		branch = os.Getenv("GITHUB_HEAD_REF")
		if branch == "" {
			branch = os.Getenv("GITHUB_REF_NAME")
		}
	}
	if branch == "" {
		v.ok("nom de branche : aucun sur HEAD détachée")
	} else if compiled.MatchString(branch) {
		v.ko("nom de branche : occurrence sensible")
	} else {
		v.ok("nom de branche : aucune occurrence")
	}
}

func (v *verifier) checkSensitiveCommits(compiled *regexp.Regexp) {
	if _, err := v.command(nil, "git", "rev-parse", "origin/main"); err == nil {
		messages, logErr := v.command(nil, "git", "log", "origin/main..HEAD", "--format=%B")
		if logErr != nil {
			v.ko("messages de commit non poussés : lecture interrompue")
		} else {
			hits := 0
			for _, line := range strings.Split(string(messages), "\n") {
				if compiled.MatchString(line) {
					hits++
				}
			}
			if hits > 0 {
				v.ko(fmt.Sprintf("messages de commit non poussés : %d occurrence(s)", hits))
			} else {
				v.okIf("messages de commit non poussés : aucune occurrence")
			}
		}
	}
}

func (v *verifier) gitNullFiles(args ...string) ([]string, error) {
	output, err := v.command(nil, "git", args...)
	if err != nil {
		return nil, err
	}
	trimmed := bytes.TrimSuffix(output, []byte{0})
	if len(trimmed) == 0 {
		return nil, nil
	}
	parts := bytes.Split(trimmed, []byte{0})
	result := make([]string, len(parts))
	for index, part := range parts {
		result[index] = string(part)
	}
	return result, nil
}
