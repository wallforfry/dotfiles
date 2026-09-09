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

	if !requireDocker(runtime, "firecrawl-mcp") {
		return ExitUnavailable
	}
	if file, err := os.Open(compose); err != nil {
		fprintf(runtime.Stderr, "firecrawl-mcp: composition absente (%s) - lancer chezmoi apply\n", compose)
		return ExitUnavailable
	} else {
		_ = file.Close()
	}

	switch option {
	case "--start":
		if code := runCompose(runtime, compose, "up", "--detach", "--wait", "--wait-timeout", timeout); code != 0 {
			return code
		}
		fprintf(runtime.Stdout, "🔥  Firecrawl écoute sur %s\n", apiURL)
	case "--stop":
		if code := runCompose(runtime, compose, "down"); code != 0 {
			return code
		}
		fprintf(runtime.Stdout, "🔥  Firecrawl arrêté\n")
	case "--status":
		return runCompose(runtime, compose, "ps")
	case "":
		if !firecrawlResponds(runtime, apiURL) {
			process := composeProcess(runtime, compose, "up", "--detach", "--wait", "--wait-timeout", timeout)
			process.Stdout = runtime.Stderr
			if runtime.Executor.Run(process) != nil {
				fprintf(runtime.Stderr, "firecrawl-mcp: pile non démarrée - lancer « firecrawl-mcp --start » une fois\n")
				return ExitUnavailable
			}
		}
		process := runtime.process("npx", "--yes", "firecrawl-mcp")
		process.Env = replaceEnv(process.Env, "FIRECRAWL_API_URL", apiURL)
		process.Replace = true
		return exitCode(runtime.Executor.Run(process))
	default:
		fprintf(runtime.Stderr, "firecrawl-mcp: option inconnue « %s » (--start, --stop, --status)\n", option)
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

func firecrawlResponds(runtime Runtime, apiURL string) bool {
	return runtime.Executor.Run(runtime.silent("curl", "-fsS", "--max-time", "2", apiURL+"/v0/health/liveness")) == nil
}

func composeProcess(runtime Runtime, compose string, args ...string) Process {
	allArgs := append([]string{"compose", "--file", compose}, args...)
	return runtime.process("docker", allArgs...)
}

func runCompose(runtime Runtime, compose string, args ...string) int {
	return exitCode(runtime.Executor.Run(composeProcess(runtime, compose, args...)))
}

func replaceEnv(environment []string, key, value string) []string {
	prefix := key + "="
	replaced := make([]string, 0, len(environment)+1)
	for _, entry := range environment {
		if len(entry) < len(prefix) || entry[:len(prefix)] != prefix {
			replaced = append(replaced, entry)
		}
	}
	return append(replaced, prefix+value)
}
