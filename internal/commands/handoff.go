package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type hookInput struct {
	TranscriptPath string `json:"transcript_path"`
	SessionID      string `json:"session_id"`
	StopHookActive bool   `json:"stop_hook_active"`
}

type transcriptEntry struct {
	Type        string `json:"type"`
	IsSidechain bool   `json:"isSidechain"`
	Message     struct {
		Usage *struct {
			InputTokens              int64 `json:"input_tokens"`
			CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
			CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

type hookDecision struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

func AgentHandoff(runtime Runtime, _ []string) int {
	var input hookInput
	if json.NewDecoder(runtime.Stdin).Decode(&input) != nil || input.StopHookActive {
		return 0
	}
	if input.TranscriptPath == "" {
		return 0
	}
	transcript, err := readLastLines(input.TranscriptPath, 500)
	if err != nil {
		return 0
	}

	sentinel := filepath.Join(handoffStateDir(runtime), handoffSession(input.SessionID))
	if _, err := os.Stat(sentinel); err == nil {
		return 0
	}

	used, present := lastTokenUsage(transcript)
	if !present {
		return 0
	}
	threshold, valid := handoffThreshold(runtime)
	if !valid || used < threshold {
		return 0
	}
	return blockHandoff(runtime, sentinel, used, threshold)
}

func handoffSession(session string) string {
	if session == "" {
		return "unknown"
	}
	return session
}

func blockHandoff(runtime Runtime, sentinel string, used, threshold int64) int {
	if os.MkdirAll(filepath.Dir(sentinel), 0o755) != nil {
		return 1
	}
	file, err := os.OpenFile(sentinel, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return 1
	}
	if err := file.Close(); err != nil {
		return 1
	}

	decision := hookDecision{
		Decision: "block",
		Reason: fmt.Sprintf("Context is at %dk tokens, past the %dk handoff threshold. ", used/1000, threshold/1000) +
			"Start no new work. Use /handoff to emit the resume prompt for a fresh session, then stop.",
	}
	if err := json.NewEncoder(runtime.Stdout).Encode(decision); err != nil {
		return 1
	}
	return 0
}

func readLastLines(path string, limit int) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	const blockSize int64 = 64 * 1024
	position := info.Size()
	var blocks [][]byte
	newlines := 0
	for position > 0 && newlines <= limit {
		size := min(blockSize, position)
		position -= size
		block := make([]byte, size)
		if _, err := file.ReadAt(block, position); err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
		newlines += bytes.Count(block, []byte{'\n'})
	}
	var contents bytes.Buffer
	for index := len(blocks) - 1; index >= 0; index-- {
		contents.Write(blocks[index])
	}
	lines := bytes.Split(contents.Bytes(), []byte{'\n'})
	if len(lines) > limit {
		lines = lines[len(lines)-limit-1:]
	}
	return bytes.Join(lines, []byte{'\n'}), nil
}

func handoffStateDir(runtime Runtime) string {
	if state, present := runtime.LookupEnv("XDG_STATE_HOME"); present && state != "" {
		return filepath.Join(state, "claude", "handoff")
	}
	return filepath.Join(runtime.env("HOME", ""), ".local", "state", "claude", "handoff")
}

func lastTokenUsage(contents []byte) (int64, bool) {
	lines := bytes.Split(contents, []byte{'\n'})
	if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > 500 {
		lines = lines[len(lines)-500:]
	}
	var used int64
	found := false
	for _, line := range lines {
		var entry transcriptEntry
		if json.Unmarshal(line, &entry) != nil || entry.Type != "assistant" || entry.IsSidechain || entry.Message.Usage == nil {
			continue
		}
		usage := entry.Message.Usage
		used = usage.InputTokens + usage.CacheReadInputTokens + usage.CacheCreationInputTokens
		found = true
	}
	return used, found
}

func handoffThreshold(runtime Runtime) (int64, bool) {
	if threshold, present := runtime.LookupEnv("HANDOFF_TOKEN_THRESHOLD"); present && threshold != "" {
		return positiveDecimal(threshold)
	}
	window, present := runtime.LookupEnv("CLAUDE_CODE_AUTO_COMPACT_WINDOW")
	if !present || window == "" {
		return 0, false
	}
	parsed, valid := positiveDecimal(window)
	if !valid {
		return 0, false
	}
	return parsed * 85 / 100, true
}

func positiveDecimal(value string) (int64, bool) {
	for _, character := range value {
		if character < '0' || character > '9' {
			return 0, false
		}
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	return parsed, err == nil && parsed > 0
}
