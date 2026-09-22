package commands

import (
	"os"
	"path/filepath"
)

func Firecrawl(runtime Runtime, args []string) int {
	compose := firecrawlCompose(runtime)
	apiURL := runtime.env("FIRECRAWL_API_URL", "http://localhost:3002")
	timeout := runtime.env("FIRECRAWL_WAIT_TIMEOUT", "90")
	option := firstArg(args)

	if !requireDocker(runtime, "firecrawl") {
		return ExitUnavailable
	}
	if file, err := os.Open(compose); err != nil {
		fprintf(runtime.Stderr, "firecrawl: composition absente (%s) - lancer chezmoi apply\n", compose)
		return ExitUnavailable
	} else {
		_ = file.Close()
	}

	switch option {
	case "--start":
		if code := runCompose(runtime, compose, "up", "--detach", "--wait", "--wait-timeout", timeout); code != 0 {
			return code
		}
		// L'api n'a pas de healthcheck : --wait rend la main quelques secondes
		// avant qu'elle n'écoute, et le premier appel échouerait.
		if !firecrawlListens(runtime, apiURL, timeout) {
			fprintf(runtime.Stderr, "firecrawl: l'API ne répond pas sur %s après %s s\n", apiURL, timeout)
			return ExitUnavailable
		}
		fprintf(runtime.Stdout, "🔥  Firecrawl écoute sur %s\n", apiURL)
	case "--stop":
		if code := runCompose(runtime, compose, "down"); code != 0 {
			return code
		}
		fprintf(runtime.Stdout, "🔥  Firecrawl arrêté\n")
	case "--status":
		return runCompose(runtime, compose, "ps")
	default:
		fprintf(runtime.Stderr, "firecrawl: usage - --start, --stop, --status\n")
		return ExitUsage
	}
	return 0
}

func firecrawlCompose(runtime Runtime) string {
	if compose, present := runtime.LookupEnv("FIRECRAWL_COMPOSE"); present && compose != "" {
		return compose
	}
	home := runtime.env("HOME", "")
	return filepath.Join(home, ".config", "firecrawl", "compose.yml")
}

func firecrawlListens(runtime Runtime, apiURL, timeout string) bool {
	probe := runtime.silent("curl", "-fsS", "--output", "/dev/null", "--max-time", "2",
		"--retry", timeout, "--retry-delay", "1", "--retry-all-errors", "--retry-max-time", timeout,
		apiURL+"/v0/health/liveness")
	return runtime.Executor.Run(probe) == nil
}

func runCompose(runtime Runtime, compose string, args ...string) int {
	allArgs := append([]string{"compose", "--file", compose}, args...)
	return exitCode(runtime.Executor.Run(runtime.process("docker", allArgs...)))
}
