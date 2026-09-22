---
name: containerized-mcp
description: >
  Run Docker-delivered MCP servers through reusable named containers. Use when registering,
  configuring or repairing a containerized MCP command. Make sure to use it whenever an MCP setup
  would invoke `docker run` per client session, even if container lifecycle is never named.
compatibility: Requires Docker and the dotfiles Go CLI built by chezmoi.
metadata:
  category: ops
---

# Containerized MCP

## Overview

An MCP client can outlive its launcher or die without cleaning up, so a command that runs a fresh
container per session leaks containers. Put lifecycle policy in one wrapper: it starts one named
container per distinct configuration on demand, then attaches each MCP session with `docker exec`.

## Usage

Use this skill when adding or changing an MCP server distributed as a Docker image, reviewing an MCP
registration command, or diagnosing duplicate and abandoned MCP containers. Lifecycle policy lives
in `internal/commands`; deployed command names are symlinks to the single `dotfiles` binary.

## Steps

1. Decide whether the server is registered at all. A client spawns every registered server at every
   session start, used or not. A server that most sessions never call is not registered; its
   command makes one `tools/call` per invocation instead, like `scrapling` (ADR-025).
2. Identify the server's distinct runtime configuration, image, stdio command, environment inputs,
   stop operation and status operation.
3. Derive a stable container name from the configuration without embedding credentials. Different
   configurations get different containers; concurrent sessions for one configuration share one.
4. Extend the cohesive Go command that tries `docker start`, creates the named detached container
   only when absent, then retries `docker start` if creation lost a race to another session. Attach
   the MCP stdio process with `docker exec --interactive` only after one of those paths succeeds.
5. Pass credentials through the wrapper's environment and then the container environment. Never put
   them in the registered command, container name, arguments, logs or repository.
6. Add a chezmoi symlink for the stable command name and, when registered, register that path as the
   MCP command. Never register `docker run -i --rm <image>`.
7. Exercise two consecutive client sessions and confirm they reuse one container. Exercise explicit
   status and stop operations, then inspect the client handshake.

## Gotchas

- **Using `--rm` as lifecycle management** - the container is removed only after its process exits,
  which is not guaranteed when the client dies. Reuse a detached named container instead.
- **Sharing one container across distinct configurations** - sessions can reach the wrong service or
  credentials. Derive separate non-secret names from configuration identity.
- **Putting credentials in arguments or names** - process lists and Docker metadata expose them.
  Read them from the environment and keep errors free of their values.
- **Starting a heavy stack during the MCP handshake** - client startup can time out before the stack
  is ready, and every session start pays for it. Expose a separate explicit start operation instead.
- **Registering a rarely used server** - "started on demand" then means started by every session:
  Firecrawl and Scrapling ran permanently, 102 starts for 0 calls in five days. Leave it
  unregistered and call it through its command.
- **Relying on `restart: on-failure` to keep a stack down** - after an unclean daemon or VM stop,
  containers are restored with exit code 255, which on-failure restarts. Use `restart: "no"`.
- **Treating a failed create as fatal** - another session may have created the same named container
  between start and run. Retry `docker start` before reporting failure.

## Constraints

- Never register `docker run -i --rm <image>` as an MCP command.
- Never register a server that most sessions never call; give it a one-shot command instead.
- Use one named container per distinct configuration and reuse it across sessions.
- Preserve the `docker start`, `docker run`, second `docker start` sequence that closes creation races.
- Keep credentials in the environment, never in command lines, names or logs.
- Provide deterministic status and stop operations for every wrapper-owned container.
- Keep lifecycle policy in the Go command rather than duplicating it in each client registration.
