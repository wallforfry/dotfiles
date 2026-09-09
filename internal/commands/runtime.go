package commands

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
			if errors.Is(err, exec.ErrNotFound) && commandExistsOnPath(process.Name) {
				return fs.ErrPermission
			}
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

func commandExistsOnPath(name string) bool {
	if strings.ContainsRune(name, filepath.Separator) {
		_, err := os.Stat(name)
		return err == nil
	}
	for _, directory := range filepath.SplitList(os.Getenv("PATH")) {
		if _, err := os.Stat(filepath.Join(directory, name)); err == nil {
			return true
		}
	}
	return false
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
	if errors.Is(err, exec.ErrNotFound) || errors.Is(err, syscall.ENOENT) {
		return 127
	}
	if errors.Is(err, fs.ErrPermission) || errors.Is(err, syscall.EACCES) {
		return 126
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
