package commands

// Scrapling adapts github.com/SebastienElet/dotfiles under its BSD 2-Clause license.
// Copyright (c) 2014 Sébastien ELET.
func Scrapling(runtime Runtime, args []string) int {
	container := runtime.env("SCRAPLING_CONTAINER", "scrapling-mcp")
	image := runtime.env("SCRAPLING_IMAGE", "pyd4vinci/scrapling:latest")
	option := firstArg(args)

	if !requireDocker(runtime, "scrapling-mcp") {
		return ExitUnavailable
	}
	switch option {
	case "--stop":
		if runtime.Executor.Run(runtime.silent("docker", "stop", container)) == nil {
			fprintf(runtime.Stdout, "🕷  Scrapling arrêté\n")
		} else {
			fprintf(runtime.Stdout, "🕷  Scrapling n'était pas démarré\n")
		}
	case "--status":
		process := runtime.process("docker", "ps", "--all", "--filter", "name=^"+container+"$", "--format", "{{.Names}}\t{{.Image}}\t{{.Status}}")
		return exitCode(runtime.Executor.Run(process))
	case "":
		if code := ensureScrapling(runtime, container, image); code != 0 {
			return code
		}
		process := runtime.process("docker", "exec", "--interactive", container, "uv", "run", "scrapling", "mcp")
		process.Replace = true
		return exitCode(runtime.Executor.Run(process))
	default:
		fprintf(runtime.Stderr, "scrapling-mcp: option inconnue « %s » (--stop, --status)\n", option)
		return ExitUsage
	}
	return 0
}

func ensureScrapling(runtime Runtime, container, image string) int {
	if runtime.Executor.Run(runtime.silent("docker", "start", container)) == nil {
		return 0
	}
	run := runtime.silent("docker", "run", "--detach", "--name", container,
		"--add-host=host.docker.internal:host-gateway", "--volume", "scrapling-profiles:/profiles",
		"--entrypoint", "sleep", image, "infinity")
	if runtime.Executor.Run(run) == nil {
		return 0
	}
	return exitCode(runtime.Executor.Run(runtime.quietStdout("docker", "start", container)))
}
