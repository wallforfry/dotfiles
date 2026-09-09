package audit

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func captureWorktree(source, repository, scratch string) (string, error) {
	patch, err := gitOutput(source, "diff", "--binary", "HEAD")
	if err != nil {
		return "", err
	}
	if len(patch) > 0 {
		command := exec.Command("git", "apply")
		command.Dir = repository
		command.Stdin = bytes.NewReader(patch)
		if output, err := command.CombinedOutput(); err != nil {
			return "", fmt.Errorf("application de la capture: %w: %s", err, output)
		}
	}
	paths, err := gitOutput(source, "ls-files", "-z", "--others", "--exclude-standard")
	if err != nil {
		return "", err
	}
	if err := copyUntracked(source, repository, bytes.Split(bytes.TrimSuffix(paths, []byte{0}), []byte{0})); err != nil {
		return "", err
	}
	commands := [][]string{{"config", "user.name", "harness-audit"}, {"config", "user.email", "harness-audit@invalid"}, {"config", "gc.auto", "0"}, {"config", "maintenance.auto", "false"}, {"add", "-A"}, {"-c", "core.hooksPath=/dev/null", "commit", "--no-gpg-sign", "-qm", "test: capture audit baseline", "--allow-empty"}}
	for _, arguments := range commands {
		if _, err := gitOutput(repository, arguments...); err != nil {
			return "", err
		}
	}
	result, err := gitOutput(repository, "rev-parse", "HEAD")
	return string(bytes.TrimSpace(result)), err
}

func copyUntracked(source, destination string, paths [][]byte) error {
	for _, rawPath := range paths {
		if len(rawPath) == 0 {
			continue
		}
		path := string(rawPath)
		if err := copyEntry(filepath.Join(source, path), filepath.Join(destination, path)); err != nil {
			return err
		}
	}
	return nil
}

func copyEntry(source, destination string) error {
	info, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(source)
		if err != nil {
			return err
		}
		return os.Symlink(target, destination)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("entrée non suivie non copiable")
	}
	return copyRegular(source, destination, info.Mode())
}

func copyRegular(source, destination string, mode os.FileMode) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode.Perm())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Chmod(destination, mode.Perm())
}

func probeCaptureContract(scratch string) error {
	source := filepath.Join(scratch, "probe-source")
	destination := filepath.Join(scratch, "probe-destination")
	if err := os.MkdirAll(source, 0o755); err != nil {
		return err
	}
	if copyUntracked(source, destination, [][]byte{[]byte("missing-entry")}) == nil {
		return fmt.Errorf("une entrée absente a été ignorée")
	}
	if err := os.Symlink("target", filepath.Join(source, "link")); err != nil {
		return err
	}
	if err := copyUntracked(source, destination, [][]byte{[]byte("link")}); err != nil {
		return err
	}
	target, err := os.Readlink(filepath.Join(destination, "link"))
	if err != nil || target != "target" {
		return fmt.Errorf("lien symbolique non préservé")
	}
	return nil
}

func gitOutput(directory string, arguments ...string) ([]byte, error) {
	command := exec.Command("git", arguments...)
	command.Dir = directory
	return command.CombinedOutput()
}
