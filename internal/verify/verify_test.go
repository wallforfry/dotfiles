package verify

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunIntegration(t *testing.T) {
	if os.Getenv("VERIFY_INTEGRATION") == "" {
		t.Skip("set VERIFY_INTEGRATION=1 to run the repository barrier")
	}
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test source path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
	var stdout, stderr bytes.Buffer
	if code := Run(root, &stdout, &stderr); code != 0 {
		t.Fatalf("Run() = %d\nstdout:\n%s\nstderr:\n%s", code, stdout.String(), stderr.String())
	}
	t.Log(stdout.String())
}

func TestBootstrapIntegration(t *testing.T) {
	if os.Getenv("VERIFY_BOOTSTRAP_INTEGRATION") == "" {
		t.Skip("set VERIFY_BOOTSTRAP_INTEGRATION=1 to run the bootstrap control")
	}
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test source path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
	var stdout, stderr bytes.Buffer
	if code := RunControl(root, ControlBootstrap, &stdout, &stderr); code != 0 {
		t.Fatalf("RunControl() = %d\nstdout:\n%s\nstderr:\n%s", code, stdout.String(), stderr.String())
	}
	t.Log(stdout.String())
}

func TestParseFrontmatter(t *testing.T) {
	t.Parallel()
	content := `---
name: example
description: >-
  Use when a multiline description
  must be normalized.
metadata:
  category: "dev"
---
# Example
`
	want := frontmatter{
		name:        "example",
		description: "Use when a multiline description must be normalized.",
		category:    "dev",
	}
	got := parseFrontmatter(content)
	if got.name != want.name || got.description != want.description || got.category != want.category {
		t.Fatalf("parseFrontmatter() = %#v, want fields %#v", got, want)
	}
}

func TestLoadRoutingCasesRejectsDuplicateIdentifiers(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "cases.tsv")
	content := strings.Join([]string{
		"id\tkind\texpected\texcluded\tprompt",
		"same\tpositive\tone\t-\tFirst",
		"same\tnegative\t-\tone\tSecond",
	}, "\n")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := loadRoutingCases(path)
	if err == nil || !strings.Contains(err.Error(), "identifiants dupliqués") {
		t.Fatalf("loadRoutingCases() error = %v, want duplicate identifier", err)
	}
}

func TestBusinessIssueOpeningFixtures(t *testing.T) {
	t.Parallel()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test source path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
	fixtures := []struct {
		name    string
		wantErr string
	}{
		{name: "valid.md"},
		{name: "missing-next.md", wantErr: "Next"},
		{name: "too-many-steps.md", wantErr: "liste"},
		{name: "missing-completion-evidence.md", wantErr: "Completion evidence"},
		{name: "missing-progress.md", wantErr: "Progress"},
		{name: "missing-progress-next.md", wantErr: "Progress doit annoncer Next"},
		{name: "missing-error-correction.md", wantErr: "Correction"},
	}
	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(filepath.Join(root, "internal", "verify", "testdata", "business-issue-drafts", fixture.name))
			if err != nil {
				t.Fatal(err)
			}
			err = validateBusinessIssueRegressionFixture(string(content))
			if fixture.wantErr == "" && err != nil {
				t.Fatalf("validateBusinessIssueOpening() = %v, want nil", err)
			}
			if fixture.wantErr != "" && (err == nil || !strings.Contains(err.Error(), fixture.wantErr)) {
				t.Fatalf("validateBusinessIssueOpening() = %v, want error containing %q", err, fixture.wantErr)
			}
		})
	}
}

func TestBusinessIssueInstructionMatchesFixtureContract(t *testing.T) {
	t.Parallel()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test source path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
	content, err := os.ReadFile(filepath.Join(root, "dot_config", "agent-skills", "business-issue", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if err := validateBusinessIssueInstruction(string(content)); err != nil {
		t.Fatal(err)
	}
}

func TestLoadSensitivePatternsFailsClosed(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "sensible.txt")
	if err := os.WriteFile(path, []byte("# commentaire\n\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err := loadSensitivePatterns(path)
	if err == nil || !strings.Contains(err.Error(), "vide") {
		t.Fatalf("loadSensitivePatterns() error = %v, want empty-list failure", err)
	}
}

func TestDisallowedIPLines(t *testing.T) {
	t.Parallel()
	address := strings.Join([]string{"192", "168", "2", "1"}, ".")
	content := "localhost 127.0.0.1\npublic " + address + "\nlistener 0.0.0.0\n"
	got := disallowedIPLines(content)
	if len(got) != 1 || got[0] != "public "+address {
		t.Fatalf("disallowedIPLines() = %q", got)
	}
}

func TestAllControlsResolveToOneCheck(t *testing.T) {
	v := &verifier{}
	for _, control := range allControls() {
		if _, ok := v.check(control); !ok {
			t.Fatalf("control %q has no check", control)
		}
	}
}
