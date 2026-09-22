package commands

import (
	"encoding/json"
	"strings"
)

// Scrapling adapts github.com/SebastienElet/dotfiles under its BSD 2-Clause license.
// Copyright (c) 2014 Sébastien ELET.
func Scrapling(runtime Runtime, args []string) int {
	container := runtime.env("SCRAPLING_CONTAINER", "scrapling-mcp")
	image := runtime.env("SCRAPLING_IMAGE", "pyd4vinci/scrapling:latest")
	option := firstArg(args)

	if !requireDocker(runtime, "scrapling") {
		return ExitUnavailable
	}
	switch {
	case option == "--stop":
		if runtime.Executor.Run(runtime.silent("docker", "stop", container)) == nil {
			fprintf(runtime.Stdout, "🕷  Scrapling arrêté\n")
		} else {
			fprintf(runtime.Stdout, "🕷  Scrapling n'était pas démarré\n")
		}
	case option == "--status":
		process := runtime.process("docker", "ps", "--all", "--filter", "name=^"+container+"$", "--format", "{{.Names}}\t{{.Image}}\t{{.Status}}")
		return exitCode(runtime.Executor.Run(process))
	case option == "" || strings.HasPrefix(option, "-") || len(args) > 2:
		fprintf(runtime.Stderr, "scrapling: usage - <outil> [arguments-json], --stop, --status\n")
		return ExitUsage
	default:
		arguments, valid := scraplingArguments(args[1:])
		if !valid {
			fprintf(runtime.Stderr, "scrapling: les arguments doivent former un objet JSON\n")
			return ExitUsage
		}
		if code := ensureScrapling(runtime, container, image); code != 0 {
			return code
		}
		server := runtime.process("docker", "exec", "--interactive", container, "uv", "run", "scrapling", "mcp")
		return callTool(runtime, "scrapling", server, option, arguments)
	}
	return 0
}

func scraplingArguments(args []string) (map[string]any, bool) {
	arguments := map[string]any{}
	if len(args) == 0 {
		return arguments, true
	}
	if err := json.Unmarshal([]byte(args[0]), &arguments); err != nil || arguments == nil {
		return nil, false
	}
	return arguments, true
}

func ensureScrapling(runtime Runtime, container, image string) int {
	if runtime.Executor.Run(runtime.silent("docker", "start", container)) == nil {
		return 0
	}
	// --init : sans lui, sleep occupe le PID 1, ignore SIGTERM, et chaque
	// docker stop attend le délai complet.
	run := runtime.silent("docker", "run", "--detach", "--init", "--name", container,
		"--add-host=host.docker.internal:host-gateway", "--volume", "scrapling-profiles:/profiles",
		"--entrypoint", "sleep", image, "infinity")
	if runtime.Executor.Run(run) == nil {
		return 0
	}
	return exitCode(runtime.Executor.Run(runtime.quietStdout("docker", "start", container)))
}
