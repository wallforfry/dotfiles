package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wallforfry/dotfiles/internal/telemetry"
)

func (a *auditor) measureTelemetry() {
	title := fmt.Sprintf("Activation et adhérence (règles introduites le %s)", a.config.Since)
	a.section(title, func() error {
		started := time.Now()
		cache := telemetry.LoadCache(a.config.CachePath, a.config.Since)
		nextCache := map[string]telemetry.CacheEntry{}
		bySource := map[string]telemetry.Summary{}
		sources := map[string]int{}
		hits := 0
		for _, source := range []sourceRoot{{"claude", a.config.ClaudeProjects}, {"codex", a.config.CodexSessions}} {
			count, sourceHits, err := collectTelemetry(source, a.config.Since, cache, nextCache, bySource)
			if err != nil {
				return err
			}
			if count == 0 {
				fmt.Fprintf(a.stdout, "  %s : indisponible\n", source.name)
				continue
			}
			sources[source.name] = count
			hits += sourceHits
		}
		if len(sources) == 0 {
			return fmt.Errorf("aucun transcript lu")
		}
		if err := telemetry.SaveCache(a.config.CachePath, a.config.Since, nextCache); err != nil {
			return fmt.Errorf("cache : écriture interrompue")
		}
		total := telemetry.Merge(summaryValues(bySource))
		telemetry.PrintReport(a.stdout, sources, total, bySource, hits, int(time.Since(started).Milliseconds()), localSkillSizes(a.config.Root))
		if err := telemetry.ValidateFormats(total); err != nil {
			return err
		}
		a.ok("mesuré sur les transcripts disponibles")
		fmt.Fprintln(a.stdout, "  l'adhérence est corrélationnelle : le modèle a changé sur la même période")
		return nil
	})
}

type sourceRoot struct{ name, root string }

func collectTelemetry(source sourceRoot, since string, cache map[string]telemetry.CacheEntry, next map[string]telemetry.CacheEntry, totals map[string]telemetry.Summary) (int, int, error) {
	info, err := os.Stat(source.root)
	if err != nil || !info.IsDir() {
		return 0, 0, nil
	}
	entries, summaries, hits, err := telemetry.CollectSource(source.name, source.root, since, cache)
	if err != nil {
		return 0, 0, err
	}
	for identity, entry := range entries {
		next[identity] = entry
	}
	totals[source.name] = telemetry.Merge(summaries)
	return len(summaries), hits, nil
}

func localSkillSizes(root string) map[string]int64 {
	result := map[string]int64{}
	paths, _ := filepath.Glob(filepath.Join(root, "dot_config", "agent-skills", "*", "SKILL.md"))
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil {
			result[filepath.Base(filepath.Dir(path))] = info.Size()
		}
	}
	return result
}

func summaryValues(values map[string]telemetry.Summary) []telemetry.Summary {
	result := make([]telemetry.Summary, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}
