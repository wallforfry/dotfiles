package verify

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type bootstrapArchitecture struct {
	machine string
	goArch  string
	ageArch string
	musl    string
}

func (v *verifier) checkBootstrap() {
	v.head("Bootstrap sans gestionnaire")
	configOK := v.checkScriptTempDir()
	rendered, err := v.renderLinuxBootstrap()
	if err != nil {
		v.ko("run_onchange_before_install-tools.sh.tmpl ne se rend pas pour Linux")
		return
	}

	tested := 0
	for _, architecture := range bootstrapArchitectures() {
		if err := v.checkOfflineBootstrap(rendered, architecture); err != nil {
			v.ko(err.Error())
			continue
		}
		tested++
	}
	if configOK && tested == len(bootstrapArchitectures()) {
		v.ok("2 architectures Linux sans gestionnaire, cache utilisateur et dégradation vérifiés")
	}
}

func bootstrapArchitectures() []bootstrapArchitecture {
	return []bootstrapArchitecture{
		{machine: "x86_64", goArch: "amd64", ageArch: "amd64", musl: "x86_64-unknown-linux-musl"},
		{machine: "aarch64", goArch: "arm64", ageArch: "arm64", musl: "aarch64-unknown-linux-musl"},
	}
}

func (v *verifier) checkScriptTempDir() bool {
	content, err := os.ReadFile(v.path(".chezmoi.toml.tmpl"))
	if err != nil {
		v.ko(".chezmoi.toml.tmpl est illisible")
		return false
	}
	output, err := v.command(content, "chezmoi", "execute-template", "--init", "--config", v.configs[0],
		"--source", v.root, "--override-data", `{"chezmoi":{"os":"linux"}}`)
	home, homeErr := os.UserHomeDir()
	expected := fmt.Sprintf("scriptTempDir = %q", filepath.Join(home, ".cache", "chezmoi"))
	if err != nil || homeErr != nil || !containsExactLine(string(output), expected) {
		v.ko(".chezmoi.toml.tmpl ne place pas les scripts sous $HOME/.cache/chezmoi")
		return false
	}
	return true
}

func (v *verifier) renderLinuxBootstrap() ([]byte, error) {
	content, err := os.ReadFile(v.path("run_onchange_before_install-tools.sh.tmpl"))
	if err != nil {
		return nil, err
	}
	return v.command(content, "chezmoi", "execute-template", "--config", v.configs[0], "--source", v.root,
		"--override-data", `{"chezmoi":{"os":"linux"},"profile":"perso","gui":false}`)
}

func (v *verifier) checkOfflineBootstrap(rendered []byte, architecture bootstrapArchitecture) error {
	scenario := filepath.Join(v.tempDir, "bootstrap-"+architecture.machine)
	fakeBin := filepath.Join(scenario, "bin")
	fakeHome := filepath.Join(scenario, "home")
	if err := os.MkdirAll(fakeBin, 0o755); err != nil {
		return fmt.Errorf("préparation du bootstrap %s impossible: %w", architecture.machine, err)
	}
	if err := os.MkdirAll(fakeHome, 0o755); err != nil {
		return fmt.Errorf("préparation du home %s impossible: %w", architecture.machine, err)
	}
	if err := installBootstrapCommands(fakeBin, architecture.machine); err != nil {
		return fmt.Errorf("préparation des commandes %s impossible: %w", architecture.machine, err)
	}

	curlLog := filepath.Join(scenario, "curl.log")
	stdout, stderr, err := runOfflineBootstrap(rendered, fakeHome, fakeBin, curlLog, architecture.machine)
	if err != nil {
		return fmt.Errorf("bootstrap Linux sans gestionnaire interrompu sur %s: %w\n%s", architecture.machine, err, indent(firstLines(stderr, 3)))
	}
	return validateOfflineBootstrap(stdout, stderr, curlLog, architecture)
}

func installBootstrapCommands(fakeBin, machine string) error {
	for _, name := range []string{"awk", "cat", "chmod", "cut", "grep", "ln", "ls", "mkdir", "mv", "rm", "sed", "sh", "tar", "unzip"} {
		path, err := exec.LookPath(name)
		if err != nil {
			return err
		}
		if err := os.Symlink(path, filepath.Join(fakeBin, name)); err != nil {
			return err
		}
	}
	realMktemp, err := exec.LookPath("mktemp")
	if err != nil {
		return err
	}
	scripts := map[string]string{
		"curl":   "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$BOOTSTRAP_CURL_LOG\"\nexit 22\n",
		"uname":  "#!/bin/sh\nprintf '%s\\n' \"$SIMULATED_ARCH\"\n",
		"id":     "#!/bin/sh\nif [ \"${1:-}\" = -u ]; then echo 0; else exit 1; fi\n",
		"mktemp": fmt.Sprintf("#!/bin/sh\ncase \" $* \" in *\" $HOME/.cache/\"*) ;; *) exit 97 ;; esac\nexec %q \"$@\"\n", realMktemp),
	}
	for name, content := range scripts {
		if err := os.WriteFile(filepath.Join(fakeBin, name), []byte(content), 0o755); err != nil {
			return err
		}
	}
	return nil
}

func runOfflineBootstrap(rendered []byte, home, path, curlLog, machine string) (string, string, error) {
	cmd := exec.Command("/bin/sh")
	cmd.Stdin = bytes.NewReader(rendered)
	cmd.Env = []string{"HOME=" + home, "PATH=" + path, "SIMULATED_ARCH=" + machine, "BOOTSTRAP_CURL_LOG=" + curlLog}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func validateOfflineBootstrap(stdout, stderr, curlLog string, architecture bootstrapArchitecture) error {
	if stdout != "" {
		return fmt.Errorf("bootstrap Linux annonce un succès malgré le réseau indisponible sur %s", architecture.machine)
	}
	for _, expected := range []string{
		"zsh absent et aucun gestionnaire de paquets connu",
		"Go non installé : le CLI dotfiles ne sera pas construit",
		"starship non installé, prompt zsh par défaut",
	} {
		if !strings.Contains(stderr, expected) {
			return fmt.Errorf("bootstrap Linux ne signale pas sa dégradation sur %s", architecture.machine)
		}
	}
	log, err := os.ReadFile(curlLog)
	if err != nil {
		return fmt.Errorf("journal réseau absent sur %s: %w", architecture.machine, err)
	}
	for _, archive := range []string{
		".linux-" + architecture.goArch + ".tar.gz",
		"linux-" + architecture.ageArch + ".tar.gz",
		architecture.musl + ".tar.gz",
	} {
		if !bytes.Contains(log, []byte(archive)) {
			return fmt.Errorf("bootstrap Linux ne sélectionne pas les archives attendues sur %s", architecture.machine)
		}
	}
	return nil
}

func containsExactLine(content, expected string) bool {
	for _, line := range strings.Split(content, "\n") {
		if line == expected {
			return true
		}
	}
	return false
}
