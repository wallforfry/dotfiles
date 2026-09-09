package verify

import (
	"fmt"
	"path/filepath"
)

func (v *verifier) checkGo() {
	v.head("Go")
	packages, err := v.command(nil, "go", "list", "./...")
	if err != nil {
		v.ko("inventaire des packages Go interrompu")
		return
	}
	count := len(nonemptyLines(string(packages)))
	if output, err := v.command(nil, "go", "test", "./..."); err != nil {
		v.ko("tests Go rouges")
		fmt.Fprint(v.stderr, indent(firstLines(string(output), 12)))
	}
	if output, err := v.command(nil, "go", "vet", "./..."); err != nil {
		v.ko("analyse statique Go rouge")
		fmt.Fprint(v.stderr, indent(firstLines(string(output), 12)))
	}
	builds := v.crossBuilds()
	v.okIf(fmt.Sprintf("%d packages, tests et analyse statique verts, %d/4 builds", count, builds))
}

func (v *verifier) crossBuilds() int {
	targets := [][2]string{{"darwin", "amd64"}, {"darwin", "arm64"}, {"linux", "amd64"}, {"linux", "arm64"}}
	passed := 0
	for _, target := range targets {
		output := filepath.Join(v.tempDir, "dotfiles-"+target[0]+"-"+target[1])
		environment := []string{"CGO_ENABLED=0", "GOOS=" + target[0], "GOARCH=" + target[1]}
		if result, err := v.commandEnv(environment, "go", "build", "-trimpath", "-o", output, "./cmd/dotfiles"); err != nil {
			v.ko(fmt.Sprintf("build Go %s/%s rouge\n%s", target[0], target[1], indent(firstLines(string(result), 5))))
			continue
		}
		passed++
	}
	return passed
}
