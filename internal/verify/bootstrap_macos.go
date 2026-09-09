package verify

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (v *verifier) checkDarwinGoBootstrap() error {
	scenario := filepath.Join(v.tempDir, "bootstrap-darwin")
	fakeBin := filepath.Join(scenario, "bin")
	home := filepath.Join(scenario, "home")
	brewPrefix := filepath.Join(scenario, "brew-go")
	for _, directory := range []string{fakeBin, home, filepath.Join(brewPrefix, "bin")} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return fmt.Errorf("préparation du bootstrap macOS impossible: %w", err)
		}
	}
	brewLog := filepath.Join(scenario, "brew.log")
	buildLog := filepath.Join(scenario, "build.log")
	if err := installDarwinCommands(fakeBin, brewPrefix, brewLog, buildLog); err != nil {
		return fmt.Errorf("préparation des commandes macOS impossible: %w", err)
	}
	install, err := v.renderForOS("run_onchange_before_install-tools.sh.tmpl", "darwin")
	if err != nil {
		return fmt.Errorf("bootstrap macOS non rendu")
	}
	_, stderr, err := runBootstrapScript(install, home, fakeBin, brewPrefix, brewLog, buildLog)
	if err != nil {
		return fmt.Errorf("installation Go macOS interrompue: %w", err)
	}
	brewCalls, _ := os.ReadFile(brewLog)
	if !bytes.Contains(brewCalls, []byte("install go\n")) || strings.Contains(stderr, "Go compatible absent") {
		return fmt.Errorf("Go Homebrew non installé face à un ancien Go sur macOS")
	}

	build, err := v.renderForOS("run_onchange_after_build-dotfiles.sh.tmpl", "darwin")
	if err != nil {
		return fmt.Errorf("construction Go macOS non rendue")
	}
	if _, _, err := runBootstrapScript(build, home, fakeBin, brewPrefix, brewLog, buildLog); err != nil {
		return fmt.Errorf("construction Go macOS interrompue: %w", err)
	}
	used, _ := os.ReadFile(buildLog)
	if string(used) != "brew-go\n" {
		return fmt.Errorf("la construction ne sélectionne pas le Go Homebrew compatible")
	}
	return nil
}

func (v *verifier) renderForOS(name, operatingSystem string) ([]byte, error) {
	content, err := os.ReadFile(v.path(name))
	if err != nil {
		return nil, err
	}
	data := fmt.Sprintf(`{"chezmoi":{"os":%q},"profile":"perso","gui":false}`, operatingSystem)
	return v.command(content, "chezmoi", "execute-template", "--config", v.configs[0], "--source", v.root, "--override-data", data)
}

func installDarwinCommands(fakeBin, brewPrefix, brewLog, buildLog string) error {
	for _, name := range []string{"awk", "chmod", "cut", "mkdir", "mktemp", "mv", "rm"} {
		path, err := exec.LookPath(name)
		if err != nil {
			return err
		}
		if err := os.Symlink(path, filepath.Join(fakeBin, name)); err != nil {
			return err
		}
	}
	scripts := map[string]string{
		"go":   "#!/bin/sh\n[ \"${1:-}\" = env ] && { echo go1.20.0; exit 0; }\n[ \"${1:-}\" = version ] && { echo 'go version go1.20.0 darwin/amd64'; exit 0; }\nexit 90\n",
		"brew": "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$BREW_LOG\"\n[ \"$*\" = 'list --versions go' ] && exit 1\n[ \"$*\" = '--prefix go' ] && { printf '%s\\n' \"$BREW_GO_PREFIX\"; exit 0; }\nexit 0\n",
	}
	for name, content := range scripts {
		if err := os.WriteFile(filepath.Join(fakeBin, name), []byte(content), 0o755); err != nil {
			return err
		}
	}
	compatible := "#!/bin/sh\nif [ \"${1:-}\" = env ]; then echo go1.27.1; exit 0; fi\nif [ \"${1:-}\" = version ]; then echo 'go version go1.27.1 darwin/amd64'; exit 0; fi\nout=\nwhile [ $# -gt 0 ]; do [ \"$1\" = -o ] && { out=$2; break; }; shift; done\nprintf '%s\\n' brew-go >> \"$BUILD_LOG\"\nprintf '#!/bin/sh\\nexit 0\\n' > \"$out\"\nchmod +x \"$out\"\n"
	return os.WriteFile(filepath.Join(brewPrefix, "bin", "go"), []byte(compatible), 0o755)
}

func runBootstrapScript(script []byte, home, path, brewPrefix, brewLog, buildLog string) (string, string, error) {
	command := exec.Command("/bin/sh")
	command.Stdin = bytes.NewReader(script)
	command.Env = []string{"HOME=" + home, "PATH=" + path, "BREW_GO_PREFIX=" + brewPrefix, "BREW_LOG=" + brewLog, "BUILD_LOG=" + buildLog}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	return stdout.String(), stderr.String(), err
}
