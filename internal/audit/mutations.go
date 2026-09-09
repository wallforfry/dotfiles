package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wallforfry/dotfiles/internal/verify"
)

type mutationCase struct {
	promise     string
	control     verify.Control
	expectation string
	label       string
	steps       []mutationStep
}

type mutationStep struct {
	kind    string
	path    string
	old     string
	new     string
	content string
}

func replace(path, old, new string) mutationStep {
	return mutationStep{kind: "replace", path: path, old: old, new: new}
}

func write(path, content string) mutationStep {
	return mutationStep{kind: "write", path: path, content: content}
}

func action(kind, path string) mutationStep {
	return mutationStep{kind: kind, path: path}
}

func mutationCases() []mutationCase {
	return append(repositoryMutationCases(), telemetryMutationCases()...)
}

func repositoryMutationCases() []mutationCase {
	reject := "reject"
	return []mutationCase{
		{"skill-identity", verify.ControlSkills, reject, "name différent du répertoire", []mutationStep{replace("dot_config/agent-skills/adr/SKILL.md", "name: adr", "name: adrx")}},
		{"skill-index", verify.ControlSkills, reject, "skill retirée de l'index", []mutationStep{action("remove-line-prefix", "dot_config/agent-skills/README.md|| `adr`")}},
		{"skill-category", verify.ControlSkills, reject, "metadata.category invalide", []mutationStep{replace("dot_config/agent-skills/adr/SKILL.md", "category: dev", "category: misc")}},
		{"skill-index", verify.ControlSkills, reject, "description README non dérivée", []mutationStep{replace("dot_config/agent-skills/README.md", "Write, amend or supersede an architecture decision record.", "Description divergente.")}},
		{"go-syntax", verify.ControlGo, reject, "erreur de syntaxe du vérificateur", []mutationStep{action("append", "internal/verify/catalog.go|\nfunc broken( {\n")}},
		{"handoff-confinement", verify.ControlGo, reject, "session hors du répertoire d'état", []mutationStep{replace("internal/commands/handoff.go", "return fmt.Sprintf(\"%x\", sha256.Sum256([]byte(session)))", "return session")}},
		{"register-concurrency", verify.ControlGo, reject, "modification concurrente de settings.json écrasée", []mutationStep{replace("internal/commands/register.go", "os.Link(temporaryPath, path)", "os.Rename(temporaryPath, path)")}},
		{"launcher-exit-code", verify.ControlGo, reject, "commande absente ramenée de 127 à 1", []mutationStep{replace("internal/commands/runtime.go", "return 127", "return 1")}},
		{"shell-syntax", verify.ControlSyntax, reject, "erreur de syntaxe du bootstrap", []mutationStep{action("append", "run_onchange_before_install-tools.sh.tmpl|\nif true; then\n")}},
		{"shell-boundary", verify.ControlSyntax, reject, "ancien script applicatif réintroduit", []mutationStep{write("scripts/migrate-homebrew-arm64.sh", "#!/bin/sh\nexit 0\n")}},
		{"template-rendering", verify.ControlTemplates, reject, "template invalide", []mutationStep{action("append", "dot_gitconfig.tmpl|{{ .absent.champ }}\n")}},
		{"bootstrap-temp", verify.ControlBootstrap, reject, "temporaire exécutable sous /tmp", []mutationStep{replace("run_onchange_before_install-tools.sh.tmpl", "$HOME/.cache/install-tools.XXXXXX", "/tmp/install-tools.XXXXXX")}},
		{"bootstrap-network", verify.ControlBootstrap, reject, "succès annoncé après échec réseau", []mutationStep{replace("run_onchange_before_install-tools.sh.tmpl", "if fetch \"https://starship.rs/install.sh\" \"$tmp/starship-install.sh\" &&", "if true ||")}},
		{"bootstrap-macos-go", verify.ControlBootstrap, reject, "ancien Go choisi devant Homebrew", []mutationStep{replace("run_onchange_after_build-dotfiles.sh.tmpl", "prefix=$(brew --prefix go 2>/dev/null || true)", "prefix=")}},
		{"adr-index", verify.ControlADR, reject, "ADR hors index", []mutationStep{action("copy-first", "docs/adr/001-*.md|docs/adr/099-fantome.md")}},
		{"age-encryption", verify.ControlEncryption, reject, "fragment age en clair", []mutationStep{write("age-key.txt.age", "texte en clair\n")}},
		{"ci-confidentiality", verify.ControlWorkflows, reject, "chezmoi diff en CI", []mutationStep{replace(".github/workflows/verify.yml", "chezmoi status", "chezmoi diff")}},
		{"skill-index", verify.ControlSkills, reject, "skill sans ligne d'index", []mutationStep{write("dot_config/agent-skills/fantome/SKILL.md", "---\nname: fantome\ndescription: Rien. Use when jamais.\nmetadata:\n  category: dev\n---\n# Fantome\n")}},
		{"skill-layout", verify.ControlSkills, reject, "sous-répertoire hors liste", []mutationStep{write("dot_config/agent-skills/adr/evals/x.json", "{}")}},
		{"merge-verdict-evidence", verify.ControlSkills, reject, "observation présentée comme preuve", []mutationStep{replace("dot_config/agent-skills/merge-verdict/references/cases.md", "**Evidence status:** observation only, not reproducible evidence for either approval verdict.", "**Evidence status:** approval evidence.")}},
		{"subagent-identity", verify.ControlSubagents, reject, "nom de subagent divergent", []mutationStep{replace("dot_claude/agents/dotfiles-reviewer.md", "name: dotfiles-reviewer", "name: divergent")}},
		{"live-state", verify.ControlLiveState, reject, "attribut exact sur l'état vivant", []mutationStep{write("dot_claude/exact_zzz", "")}},
		{"host-neutrality", verify.ControlProjections, reject, "import hôte dans source agnostique", []mutationStep{action("append", "harness/AGENTS.md|@SOUL.md\n")}},
		{"skill-availability", verify.ControlProjections, reject, "hôte privé d'une skill", []mutationStep{action("remove", "dot_codex/skills/symlink_adr")}},
		{"skill-source", verify.ControlProjections, reject, "lien vers un arbre étranger", []mutationStep{write("dot_claude/skills/symlink_adr", "../../.agents/skills/adr\n")}},
		{"live-state", verify.ControlProjections, reject, "lien sur tout le répertoire", []mutationStep{write("dot_codex/symlink_skills", "../.config/agent-skills\n")}},
		{"live-state", verify.ControlLiveState, reject, "racine exacte sur l'état vivant", []mutationStep{action("mkdir", "exact_dot_codex")}},
		{"host-instructions", verify.ControlProjections, reject, "harness absent de Codex", []mutationStep{replace("dot_codex/AGENTS.md.tmpl", "{{ include \"harness/USER.md\" }}\n", "")}},
		{"routing-contract", verify.ControlRouting, reject, "scénario positif absent", []mutationStep{action("remove-line-prefix", "internal/verify/testdata/skill-routing-cases.tsv|adr-positive\t")}},
		{"routing-contract", verify.ControlRouting, reject, "scénario négatif absent", []mutationStep{action("remove-line-prefix", "internal/verify/testdata/skill-routing-cases.tsv|adr-negative\t")}},
		{"routing-budget", verify.ControlRouting, reject, "description trop longue", []mutationStep{replace("dot_config/agent-skills/adr/SKILL.md", "description: >", "description: >\n  "+strings.Repeat("x", 400))}},
		{"sensitive-input", verify.ControlSensitive, reject, "liste sensible absente", []mutationStep{action("remove-list", "")}},
		{"sensitive-content", verify.ControlSensitive, reject, "contenu sensible non suivi", []mutationStep{write("ordinary.txt", "$MARKER\n")}},
		{"sensitive-read", verify.ControlSensitive, reject, "fichier non suivi illisible", []mutationStep{action("symlink", "ordinary-link|absent-target")}},
		{"sensitive-path", verify.ControlSensitive, reject, "nom de fichier sensible", []mutationStep{write("$MARKER.txt", "safe\n")}},
		{"sensitive-branch", verify.ControlSensitive, reject, "nom de branche sensible", []mutationStep{action("branch", "$MARKER")}},
		{"sensitive-commit", verify.ControlSensitive, reject, "message de commit sensible", []mutationStep{action("commit", "$MARKER")}},
		{"sensitive-path", verify.ControlSensitive, "accept", "nom de fichier ordinaire", []mutationStep{write("harness-audit-ordinary-marker.txt", "safe\n")}},
		{"sensitive-branch", verify.ControlSensitive, "accept", "nom de branche ordinaire", []mutationStep{action("branch", "harness-audit-ordinary")}},
	}
}

func telemetryMutationCases() []mutationCase {
	reject := "reject"
	return []mutationCase{
		{"telemetry-contract", verify.ControlTelemetry, reject, "format vivant non testé", []mutationStep{replace("internal/telemetry/schema.go", "\"FileChange\",", "\"FileChangeBroken\",")}},
		{"telemetry-contract", verify.ControlTelemetry, reject, "format vocal Codex non testé", []mutationStep{replace("internal/telemetry/schema.go", "\"realtime_item\",", "\"realtime_item_broken\",")}},
		{"telemetry-turn-passive", verify.ControlTelemetry, reject, "turn_aborted produit un événement", []mutationStep{replace("internal/telemetry/events.go", "case \"transcript_segment\":\n\t\treturn []event{{source: \"codex\", kind: \"text\", timestamp: record[\"timestamp\"], role: payload[\"role\"], text: payload[\"text\"]}}", "case \"transcript_segment\":\n\t\treturn []event{{source: \"codex\", kind: \"text\", timestamp: record[\"timestamp\"], role: payload[\"role\"], text: payload[\"text\"]}}\n\tcase \"turn_aborted\":\n\t\treturn []event{{source: \"codex\", kind: \"tool\", name: \"spawn_agent\"}}")}},
		{"telemetry-turn-shape", verify.ControlTelemetry, reject, "turn_aborted incomplet accepté", []mutationStep{replace("internal/telemetry/schema.go", "return stringFields(payload, \"turn_id\", \"reason\") &&\n\t\t\tintegerFields(payload, \"started_at\", \"completed_at\", \"duration_ms\")", "return stringFields(payload, \"turn_id\")")}},
		{"telemetry-counts", verify.ControlTelemetry, reject, "enregistrements lus comptés comme reconnus", []mutationStep{replace("internal/telemetry/telemetry.go", "increment(summary.RecognizedRecords, source, 1)", "increment(summary.ReadRecords, source, 1)")}},
		{"telemetry-format-error", verify.ControlTelemetry, reject, "dérive de format confondue avec lecture", []mutationStep{replace("internal/telemetry/errors.go", "formats non mesurés : %d inconnus, %d invalides", "%s : lecture interrompue")}},
		{"telemetry-read-error", verify.ControlTelemetry, reject, "lecture interrompue confondue avec format", []mutationStep{replace("internal/telemetry/errors.go", "%s : lecture interrompue", "formats non mesurés : %s inconnus")}},
		{"telemetry-cache-version", verify.ControlTelemetry, reject, "ancienne version de cache acceptée", []mutationStep{replace("internal/telemetry/cache.go", "cache.Version != CacheVersion || ", "")}},
		{"telemetry-codex-label", verify.ControlTelemetry, reject, "task_name publié comme qualification", []mutationStep{replace("internal/telemetry/telemetry.go", "return \"spawn_agent\"", "return labelled(value.payload, \"task_name\")")}},
		{"telemetry-claude-label", verify.ControlTelemetry, reject, "description publiée comme qualification", []mutationStep{replace("internal/telemetry/telemetry.go", "return \"agent\"", "return labelled(value.payload, \"description\")")}},
		{"typography", verify.ControlTelemetry, "observe", "adhérence mesurée", nil},
		{"comment-policy", verify.ControlTelemetry, "observe", "adhérence mesurée", nil},
	}
}

func applyMutation(root, list, marker string, steps []mutationStep) error {
	for _, step := range steps {
		step.path = strings.ReplaceAll(step.path, "$MARKER", marker)
		step.content = strings.ReplaceAll(step.content, "$MARKER", marker)
		if err := applyStep(root, list, step); err != nil {
			return err
		}
	}
	return nil
}

func applyStep(root, list string, step mutationStep) error {
	switch step.kind {
	case "replace":
		return replaceOnce(filepath.Join(root, step.path), step.old, step.new)
	case "write":
		return writeFile(filepath.Join(root, step.path), step.content)
	case "append":
		path, value, ok := strings.Cut(step.path, "|")
		if !ok {
			return fmt.Errorf("append invalide")
		}
		return appendFile(filepath.Join(root, path), value)
	case "remove-line-prefix":
		return removeLinePrefix(root, step.path)
	case "copy-first":
		return copyFirst(root, step.path)
	case "remove":
		return os.Remove(filepath.Join(root, step.path))
	case "remove-list":
		return os.Remove(list)
	case "mkdir":
		return os.Mkdir(filepath.Join(root, step.path), 0o755)
	case "symlink":
		path, target, _ := strings.Cut(step.path, "|")
		return os.Symlink(target, filepath.Join(root, path))
	case "branch":
		_, err := gitOutput(root, "switch", "-qc", step.path)
		return err
	case "commit":
		_, err := gitOutput(root, "-c", "core.hooksPath=/dev/null", "commit", "--no-gpg-sign", "--allow-empty", "-qm", step.path)
		return err
	}
	return fmt.Errorf("mutation inconnue: %s", step.kind)
}
