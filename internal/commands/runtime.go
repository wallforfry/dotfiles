package commands

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

const (
	ExitUsage       = 64
	ExitUnavailable = 69
)

type Process struct {
	Name    string
	Args    []string
	Env     []string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Replace bool
}

type Executor interface {
	Run(Process) error
}

type Runtime struct {
	Executor  Executor
	LookupEnv func(string) (string, bool)
	Environ   func() []string
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
}

func NewRuntime() Runtime {
	return Runtime{
		Executor:  OSExecutor{},
		LookupEnv: os.LookupEnv,
		Environ:   os.Environ,
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
	}
}

type OSExecutor struct{}

func (OSExecutor) Run(process Process) error {
	if process.Replace {
		path, err := exec.LookPath(process.Name)
		if err != nil {
			return err
		}
		return syscall.Exec(path, append([]string{process.Name}, process.Args...), process.Env)
	}
	command := exec.Command(process.Name, process.Args...)
	command.Env = process.Env
	command.Stdin = process.Stdin
	command.Stdout = process.Stdout
	command.Stderr = process.Stderr
	return command.Run()
}

func (runtime Runtime) env(key, fallback string) string {
	if value, present := runtime.LookupEnv(key); present && value != "" {
		return value
	}
	return fallback
}

func (runtime Runtime) quietStdout(name string, args ...string) Process {
	process := runtime.process(name, args...)
	process.Stdout = io.Discard
	return process
}

func (runtime Runtime) process(name string, args ...string) Process {
	return Process{
		Name:   name,
		Args:   args,
		Env:    runtime.Environ(),
		Stdin:  runtime.Stdin,
		Stdout: runtime.Stdout,
		Stderr: runtime.Stderr,
	}
}

func (runtime Runtime) silent(name string, args ...string) Process {
	process := runtime.process(name, args...)
	process.Stdout = io.Discard
	process.Stderr = io.Discard
	return process
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode()
	}
	return 1
}

func requireDocker(runtime Runtime, command string) bool {
	if runtime.Executor.Run(runtime.silent("docker", "info")) == nil {
		return true
	}
	fprintf(runtime.Stderr, "%s: daemon docker indisponible\n", command)
	return false
}

func fprintf(writer io.Writer, format string, values ...any) {
	_, _ = fmt.Fprintf(writer, format, values...)
}
