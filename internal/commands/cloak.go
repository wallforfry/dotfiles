package commands

// Cloak adapts github.com/SebastienElet/dotfiles under its BSD 2-Clause license.
// Copyright (c) 2014 Sébastien ELET.
func Cloak(runtime Runtime, args []string) int {
	container := runtime.env("CLOAK_CONTAINER", "cloak")
	image := runtime.env("CLOAK_IMAGE", "cloakhq/cloakbrowser:0.5.3")
	port := runtime.env("CLOAK_PORT", "9222")
	idle := runtime.env("CLOAK_IDLE_TIMEOUT", "300")
	option := firstArg(args)

	switch option {
	case "--start":
		if !requireDocker(runtime, "cloak") {
			return ExitUnavailable
		}
		if code := startCloak(runtime, container, image, port, idle); code != 0 {
			return code
		}
		fprintf(runtime.Stdout, "🥷  CloakBrowser en écoute, cdp_url http://host.docker.internal:%s\n", port)
	case "--stop":
		if !requireDocker(runtime, "cloak") {
			return ExitUnavailable
		}
		if runtime.Executor.Run(runtime.silent("docker", "stop", container)) == nil {
			fprintf(runtime.Stdout, "🥷  CloakBrowser arrêté\n")
		} else {
			fprintf(runtime.Stdout, "🥷  CloakBrowser n'était pas démarré\n")
		}
	case "--status":
		if !requireDocker(runtime, "cloak") {
			return ExitUnavailable
		}
		return exitCode(runtime.Executor.Run(runtime.process("docker", "ps", "--all", "--filter", "name=^"+container+"$", "--format", "{{.Names}}\t{{.Image}}\t{{.Status}}")))
	case "--url":
		fprintf(runtime.Stdout, "http://host.docker.internal:%s\n", port)
	default:
		fprintf(runtime.Stderr, "cloak: usage - --start, --stop, --status, --url\n")
		return ExitUsage
	}
	return 0
}

func startCloak(runtime Runtime, container, image, port, idle string) int {
	if runtime.Executor.Run(runtime.silent("docker", "start", container)) == nil {
		return 0
	}
	process := runtime.quietStdout("docker", "run", "--detach", "--name", container,
		"--publish", "127.0.0.1:"+port+":"+port, image, "cloakserve", "--idle-timeout="+idle)
	if runtime.Executor.Run(process) == nil {
		return 0
	}
	return exitCode(runtime.Executor.Run(runtime.quietStdout("docker", "start", container)))
}

func firstArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}
