package verify

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Run executes the repository's mechanical verification barrier.
func Run(root string, stdout, stderr io.Writer) int {
	return runControls(root, stdout, stderr, allControls())
}

// RunControl executes one named section of the mechanical barrier.
func RunControl(root string, control Control, stdout, stderr io.Writer) int {
	return runControls(root, stdout, stderr, []Control{control})
}

func runControls(root string, stdout, stderr io.Writer, controls []Control) int {
	v := &verifier{root: root, stdout: stdout, stderr: stderr}
	if err := v.prepare(); err != nil {
		fmt.Fprintf(stderr, "❌  %s\n", err)
		return 1
	}
	defer os.RemoveAll(v.tempDir)

	for _, control := range controls {
		check, ok := v.check(control)
		if !ok {
			v.ko(fmt.Sprintf("contrôle Go inconnu : %s", control))
			continue
		}
		check()
	}
	fmt.Fprintln(stdout)
	if v.failed {
		fmt.Fprintln(stderr, "💥  barrière rouge")
		return 1
	}
	fmt.Fprintln(stdout, "🎉  barrière verte")
	return 0
}

type verifier struct {
	root          string
	stdout        io.Writer
	stderr        io.Writer
	tempDir       string
	configs       []string
	failed        bool
	sectionFailed bool
}

func (v *verifier) prepare() error {
	root, err := filepath.Abs(v.root)
	if err != nil {
		return fmt.Errorf("racine invalide: %w", err)
	}
	v.root = root
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("répertoire personnel introuvable: %w", err)
	}
	live := os.Getenv("CHEZMOI_CONFIG")
	if live == "" {
		live = filepath.Join(home, ".config", "chezmoi", "chezmoi.toml")
	}
	base, err := os.ReadFile(live)
	if err != nil {
		return fmt.Errorf("%s absent ou illisible : lancer chezmoi init d'abord", live)
	}
	cache := filepath.Join(home, ".cache")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		return fmt.Errorf("création de %s impossible: %w", cache, err)
	}
	v.tempDir, err = os.MkdirTemp(cache, "verify.")
	if err != nil {
		return fmt.Errorf("création du répertoire temporaire impossible: %w", err)
	}
	profileLine := regexp.MustCompile(`(?m)^    profile = .*`)
	guiLine := regexp.MustCompile(`(?m)^    gui = .*`)
	for _, combo := range []struct {
		profile string
		gui     bool
	}{{"pro", true}, {"perso", true}, {"perso", false}} {
		content := profileLine.ReplaceAll(base, []byte(fmt.Sprintf(`    profile = %q`, combo.profile)))
		content = guiLine.ReplaceAll(content, []byte(fmt.Sprintf("    gui = %t", combo.gui)))
		path := filepath.Join(v.tempDir, fmt.Sprintf("%s-%t.toml", combo.profile, combo.gui))
		if err := os.WriteFile(path, content, 0o600); err != nil {
			return fmt.Errorf("écriture de %s impossible: %w", path, err)
		}
		v.configs = append(v.configs, path)
	}
	return nil
}

func (v *verifier) head(title string) {
	v.sectionFailed = false
	fmt.Fprintf(v.stdout, "\n== %s\n", title)
}

func (v *verifier) ok(message string) { fmt.Fprintf(v.stdout, "  ✅  %s\n", message) }

func (v *verifier) ko(message string) {
	v.failed = true
	v.sectionFailed = true
	fmt.Fprintf(v.stderr, "  ❌  %s\n", message)
}

func (v *verifier) okIf(message string) {
	if !v.sectionFailed {
		v.ok(message)
	}
}

func (v *verifier) command(stdin []byte, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = v.root
	cmd.Stdin = bytes.NewReader(stdin)
	cmd.Env = withoutIntegrationFlags(os.Environ())
	return cmd.CombinedOutput()
}

func withoutIntegrationFlags(environment []string) []string {
	result := make([]string, 0, len(environment))
	for _, variable := range environment {
		if !strings.HasPrefix(variable, "VERIFY_INTEGRATION=") &&
			!strings.HasPrefix(variable, "VERIFY_BOOTSTRAP_INTEGRATION=") &&
			!strings.HasPrefix(variable, "AUDIT_INTEGRATION=") {
			result = append(result, variable)
		}
	}
	return result
}

func (v *verifier) commandEnv(environment []string, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = v.root
	cmd.Env = mergeEnvironment(withoutIntegrationFlags(os.Environ()), environment)
	return cmd.CombinedOutput()
}

func mergeEnvironment(base, overrides []string) []string {
	result := append([]string{}, base...)
	for _, override := range overrides {
		key := strings.SplitN(override, "=", 2)[0] + "="
		filtered := result[:0]
		for _, variable := range result {
			if !strings.HasPrefix(variable, key) {
				filtered = append(filtered, variable)
			}
		}
		result = append(filtered, override)
	}
	return result
}

func (v *verifier) gitLines(args ...string) ([]string, error) {
	output, err := v.command(nil, "git", args...)
	if err != nil {
		return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return nonemptyLines(string(output)), nil
}

func nonemptyLines(value string) []string {
	lines := strings.Split(strings.TrimSuffix(value, "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	return lines
}

func read(path string) ([]byte, error) { return os.ReadFile(path) }

func (v *verifier) path(parts ...string) string {
	return filepath.Join(append([]string{v.root}, parts...)...)
}

func glob(pattern string) []string {
	matches, _ := filepath.Glob(pattern)
	return matches
}

func globDirs(pattern string) []string {
	var directories []string
	for _, match := range glob(pattern) {
		if info, err := os.Stat(match); err == nil && info.IsDir() {
			directories = append(directories, match)
		}
	}
	return directories
}

func firstLines(value string, count int) string {
	lines := nonemptyLines(value)
	if len(lines) > count {
		lines = lines[:count]
	}
	return strings.Join(lines, "\n")
}

func indent(value string) string {
	if value == "" {
		return ""
	}
	return "      " + strings.ReplaceAll(value, "\n", "\n      ") + "\n"
}
