package audit

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRunRequiresExplicitRoot(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run(Config{}, &stdout, &stderr); code != 1 {
		t.Fatalf("Run() = %d, want 1", code)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("racine du dépôt absente")) {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestProbeCaptureContract(t *testing.T) {
	if err := probeCaptureContract(t.TempDir()); err != nil {
		t.Fatal(err)
	}
}

func TestCopyUntrackedPreservesModeAndSymlink(t *testing.T) {
	source := t.TempDir()
	destination := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "tool"), []byte("x"), 0o751); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("tool", filepath.Join(source, "link")); err != nil {
		t.Fatal(err)
	}
	if err := copyUntracked(source, destination, [][]byte{[]byte("tool"), []byte("link")}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(destination, "tool"))
	if err != nil || info.Mode().Perm() != 0o751 {
		t.Fatalf("copied mode = %v, error = %v", info.Mode().Perm(), err)
	}
	target, err := os.Readlink(filepath.Join(destination, "link"))
	if err != nil || target != "tool" {
		t.Fatalf("copied link = %q, error = %v", target, err)
	}
}

func TestMutationMatrixKeepsPromiseCounts(t *testing.T) {
	counts := map[string]int{}
	for _, item := range mutationCases() {
		counts[item.expectation]++
		if item.promise == "" || item.control == "" || item.label == "" {
			t.Fatalf("incomplete mutation: %#v", item)
		}
	}
	if counts["reject"] != 43 || counts["accept"] != 2 || counts["observe"] != 2 {
		t.Fatalf("matrix counts = %#v", counts)
	}
}

func TestApplyMutationExpandsMarkerWithoutPersistingPath(t *testing.T) {
	root := t.TempDir()
	list := filepath.Join(t.TempDir(), "sensible.txt")
	step := write("$MARKER.txt", "$MARKER\n")
	if err := applyMutation(root, list, "secret-marker", []mutationStep{step}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(root, "secret-marker.txt"))
	if err != nil || string(content) != "secret-marker\n" {
		t.Fatalf("content = %q, error = %v", content, err)
	}
}

func TestMutationMatrixIntegration(t *testing.T) {
	if os.Getenv("AUDIT_INTEGRATION") == "" {
		t.Skip("set AUDIT_INTEGRATION=1 to run the mutation matrix")
	}
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test source path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
	var stdout, stderr bytes.Buffer
	a := auditor{config: Config{Root: root}, stdout: &stdout, stderr: &stderr}
	if err := a.runMutationMatrix(t.TempDir()); err != nil {
		t.Fatalf("matrix failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}
	t.Log(stdout.String())
}
