package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"
)

const (
	mcpProtocolVersion = "2025-06-18"
	initializeID       = 1
	callID             = 2
)

type rpcMessage struct {
	ID     *int            `json:"id"`
	Method string          `json:"method"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type toolResult struct {
	Content []struct {
		Type     string `json:"type"`
		Text     string `json:"text"`
		Data     string `json:"data"`
		MimeType string `json:"mimeType"`
	} `json:"content"`
	IsError bool `json:"isError"`
}

func callTool(runtime Runtime, command string, server Process, tool string, arguments map[string]any) int {
	input, requests := io.Pipe()
	session := &mcpSession{requests: requests, tool: tool, arguments: arguments}
	server.Stdin = input
	server.Stdout = session
	// L'entrée reste ouverte jusqu'à la réponse : un serveur mort avant de
	// répondre bloquerait sinon Wait sur ce tube.
	server.WaitDelay = 2 * time.Second
	go session.send(map[string]any{
		"jsonrpc": "2.0", "id": initializeID, "method": "initialize",
		"params": map[string]any{
			"protocolVersion": mcpProtocolVersion,
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "dotfiles-" + command, "version": "1"},
		},
	})
	err := runtime.Executor.Run(server)
	_ = requests.Close()

	switch {
	case session.response == nil:
		fprintf(runtime.Stderr, "%s: le serveur MCP s'est arrêté sans répondre\n", command)
		if code := exitCode(err); code != 0 {
			return code
		}
		return 1
	case session.response.Error != nil:
		fprintf(runtime.Stderr, "%s: %s\n", command, session.response.Error.Message)
		return 1
	}
	return printToolResult(runtime, command, session.response.Result)
}

type mcpSession struct {
	requests  *io.PipeWriter
	tool      string
	arguments map[string]any
	pending   []byte
	response  *rpcMessage
}

func (session *mcpSession) Write(chunk []byte) (int, error) {
	session.pending = append(session.pending, chunk...)
	for {
		end := bytes.IndexByte(session.pending, '\n')
		if end < 0 {
			return len(chunk), nil
		}
		session.receive(session.pending[:end])
		session.pending = session.pending[end+1:]
	}
}

func (session *mcpSession) receive(line []byte) {
	var message rpcMessage
	if json.Unmarshal(line, &message) != nil || message.ID == nil || message.Method != "" {
		return
	}
	switch {
	case *message.ID == initializeID && message.Error == nil:
		go session.send(
			map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"},
			map[string]any{
				"jsonrpc": "2.0", "id": callID, "method": "tools/call",
				"params": map[string]any{"name": session.tool, "arguments": session.arguments},
			},
		)
	case *message.ID == initializeID || *message.ID == callID:
		session.response = &message
		_ = session.requests.Close()
	}
}

func (session *mcpSession) send(messages ...map[string]any) {
	for _, message := range messages {
		encoded, _ := json.Marshal(message)
		if _, err := session.requests.Write(append(encoded, '\n')); err != nil {
			return
		}
	}
}

func printToolResult(runtime Runtime, command string, raw json.RawMessage) int {
	var result toolResult
	if err := json.Unmarshal(raw, &result); err != nil {
		fprintf(runtime.Stderr, "%s: réponse d'outil illisible: %v\n", command, err)
		return 1
	}
	output := runtime.Stdout
	if result.IsError {
		output = runtime.Stderr
	}
	for _, content := range result.Content {
		switch content.Type {
		case "text":
			fprintf(output, "%s\n", strings.TrimRight(content.Text, "\n"))
		case "image":
			path, err := saveImage(content.Data, content.MimeType)
			if err != nil {
				fprintf(runtime.Stderr, "%s: image non enregistrée: %v\n", command, err)
				return 1
			}
			fprintf(output, "%s\n", path)
		default:
			fprintf(runtime.Stderr, "%s: contenu %q ignoré\n", command, content.Type)
		}
	}
	if result.IsError {
		return 1
	}
	return 0
}

func saveImage(data, mimeType string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	extension := mimeType[strings.LastIndex(mimeType, "/")+1:]
	file, err := os.CreateTemp("", "mcp-*."+extension)
	if err != nil {
		return "", err
	}
	if _, err := file.Write(decoded); err != nil {
		_ = file.Close()
		return "", err
	}
	return file.Name(), file.Close()
}
