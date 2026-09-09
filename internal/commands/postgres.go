package commands

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
)

func Postgres(runtime Runtime, args []string) int {
	image := runtime.env("POSTGRES_MCP_IMAGE", "crystaldba/postgres-mcp")
	accessMode := runtime.env("POSTGRES_MCP_ACCESS_MODE", "--access-mode=restricted")
	option := firstArg(args)

	if !requireDocker(runtime, "postgres-mcp") {
		return ExitUnavailable
	}
	if option == "--status" {
		process := runtime.process("docker", "ps", "--all", "--filter", "name=^postgres-mcp-", "--format", "{{.Names}}\t{{.Image}}\t{{.Status}}")
		return exitCode(runtime.Executor.Run(process))
	}
	if option != "" && option != "--stop" {
		fprintf(runtime.Stderr, "postgres-mcp: option inconnue « %s » (--stop, --status)\n", option)
		return ExitUsage
	}

	uri := runtime.env("DATABASE_URI", "")
	if uri == "" {
		fprintf(runtime.Stderr, "postgres-mcp: DATABASE_URI absente de l'environnement\n")
		return ExitUsage
	}
	warnLocalPostgres(runtime, uri)
	container := postgresContainer(uri)
	if option == "--stop" {
		if runtime.Executor.Run(runtime.silent("docker", "stop", container)) == nil {
			fprintf(runtime.Stdout, "🐘  %s arrêté\n", container)
		} else {
			fprintf(runtime.Stdout, "🐘  %s n'était pas démarré\n", container)
		}
		return 0
	}

	if code := ensurePostgres(runtime, container, image); code != 0 {
		return code
	}
	process := runtime.process("docker", "exec", "--interactive", container, "postgres-mcp", accessMode)
	process.Replace = true
	return exitCode(runtime.Executor.Run(process))
}

func postgresContainer(uri string) string {
	digest := sha1.Sum([]byte(uri))
	return "postgres-mcp-" + hex.EncodeToString(digest[:])[:8]
}

func warnLocalPostgres(runtime Runtime, uri string) {
	if strings.Contains(uri, "@localhost:") || strings.Contains(uri, "@127.0.0.1:") {
		fprintf(runtime.Stderr, "postgres-mcp: DATABASE_URI pointe sur localhost, injoignable depuis un conteneur - utiliser host.docker.internal\n")
	}
}

func ensurePostgres(runtime Runtime, container, image string) int {
	if runtime.Executor.Run(runtime.silent("docker", "start", container)) == nil {
		return 0
	}
	run := runtime.silent("docker", "run", "--detach", "--name", container,
		"--add-host=host.docker.internal:host-gateway", "--env", "DATABASE_URI",
		"--entrypoint", "sleep", image, "infinity")
	if runtime.Executor.Run(run) == nil {
		return 0
	}
	return exitCode(runtime.Executor.Run(runtime.quietStdout("docker", "start", container)))
}
