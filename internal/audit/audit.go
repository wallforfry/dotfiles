package audit

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Root           string
	ClaudeProjects string
	CodexSessions  string
	Since          string
	CachePath      string
}

type auditor struct {
	config Config
	stdout io.Writer
	stderr io.Writer
	failed bool
}

// Run measures the harness and returns a process exit code.
func Run(config Config, stdout, stderr io.Writer) int {
	a := &auditor{config: config, stdout: stdout, stderr: stderr}
	if err := a.prepare(); err != nil {
		fmt.Fprintf(stderr, "❌  %s\n", err)
		return 1
	}
	started := time.Now()
	a.measureContext()
	a.measureDeployment()
	a.measureTelemetry()
	a.measureBarrier()
	fmt.Fprintf(stdout, "  Durée totale : %d ms\n\n", time.Since(started).Milliseconds())
	if a.failed {
		fmt.Fprintln(stderr, "💥  une mesure manque ou la barrière laisse passer un défaut")
		return 1
	}
	fmt.Fprintln(stdout, "🎉  mesures complètes")
	return 0
}

func (a *auditor) prepare() error {
	if a.config.Root == "" {
		return fmt.Errorf("racine du dépôt absente : aucune mesure possible")
	}
	root, err := filepath.Abs(a.config.Root)
	if err != nil {
		return fmt.Errorf("racine invalide: %w", err)
	}
	a.config.Root = root
	if !gitRepository(root) {
		return fmt.Errorf("racine hors d'un dépôt git : aucune mesure possible")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("répertoire personnel introuvable: %w", err)
	}
	applyDefaults(&a.config, home)
	return nil
}

func applyDefaults(config *Config, home string) {
	if config.ClaudeProjects == "" {
		config.ClaudeProjects = filepath.Join(home, ".claude", "projects")
	}
	if config.CodexSessions == "" {
		config.CodexSessions = filepath.Join(home, ".codex", "sessions")
	}
	if config.Since == "" {
		config.Since = "2026-08-17"
	}
	if config.CachePath == "" {
		config.CachePath = filepath.Join(home, ".cache", "harness-audit", "telemetry-v1.json")
	}
}

func (a *auditor) section(title string, measure func() error) {
	started := time.Now()
	fmt.Fprintf(a.stdout, "\n== %s\n", title)
	if err := measure(); err != nil {
		a.failed = true
		fmt.Fprintf(a.stderr, "  ❌  %s\n", err)
	}
	fmt.Fprintf(a.stdout, "  Durée : %d ms\n", time.Since(started).Milliseconds())
}

func (a *auditor) ok(format string, arguments ...any) {
	fmt.Fprintf(a.stdout, "  ✅  "+format+"\n", arguments...)
}
