package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const hindsightPackage = "@vectorize-io/hindsight-coding-agents"

var nonServerName = regexp.MustCompile(`[^a-z0-9]+`)

type hindsightRegistration struct {
	Repository string `json:"repository"`
	Bank       string `json:"bank"`
}

type hindsightConfiguration struct {
	APIURL        string                  `json:"apiUrl"`
	APIToken      string                  `json:"apiToken"`
	Registrations []hindsightRegistration `json:"registrations"`
}

type hindsightManaged struct {
	Repositories []string `json:"repositories"`
	Banks        []string `json:"banks"`
}

func RegisterHindsight(runtime Runtime, args []string) int {
	configuration, err := parseHindsightConfiguration(args)
	if err != nil {
		fprintf(runtime.Stderr, "dotfiles: %s\n", err)
		fprintf(runtime.Stderr, "dotfiles: usage - register-hindsight --config <fichier JSON chiffré déployé>\n")
		return ExitUsage
	}
	if err := registerHindsightFiles(runtime.env("HOME", ""), configuration); err != nil {
		fprintf(runtime.Stderr, "dotfiles: configuration Hindsight inchangée : %s\n", err)
		return 1
	}
	process := runtime.process("bunx", "--yes", hindsightPackage, "install", "claude-code", "codex")
	process.Env = append(process.Env, "HINDSIGHT_API_TOKEN="+configuration.APIToken)
	if err := runtime.Executor.Run(process); err != nil {
		fprintf(runtime.Stderr, "dotfiles: intégrations Claude Code et Codex non installées : %s\n", err)
		return exitCode(err)
	}
	if _, err := exec.LookPath("hindsight"); err == nil {
		configure := runtime.process("hindsight", "configure", "--api-url", configuration.APIURL, "--api-key", configuration.APIToken)
		configure.Stdout = io.Discard
		configure.Stderr = io.Discard
		if err := runtime.Executor.Run(configure); err != nil {
			fprintf(runtime.Stderr, "dotfiles: CLI Hindsight non configurée : %s\n", err)
			return exitCode(err)
		}
	}
	fprintf(runtime.Stdout, "Hindsight enregistré pour %d dépôt(s), Claude Code, Codex et Cursor. Ajoute l'application MCP dans ChatGPT depuis Réglages > Apps > Créer.\n", len(configuration.Registrations))
	return 0
}

func parseHindsightConfiguration(args []string) (hindsightConfiguration, error) {
	if len(args) != 2 || args[0] != "--config" || args[1] == "" {
		return hindsightConfiguration{}, errors.New("--config est requis")
	}
	content, err := os.ReadFile(args[1])
	if err != nil {
		return hindsightConfiguration{}, err
	}
	var configuration hindsightConfiguration
	if err := json.Unmarshal(content, &configuration); err != nil {
		return hindsightConfiguration{}, errors.New("JSON Hindsight invalide")
	}
	configuration.APIURL = strings.TrimRight(configuration.APIURL, "/")
	if configuration.APIURL == "" || configuration.APIToken == "" {
		return hindsightConfiguration{}, errors.New("apiUrl et apiToken sont requis")
	}
	seenRepositories := map[string]bool{}
	for index := range configuration.Registrations {
		registration := &configuration.Registrations[index]
		if registration.Repository == "" || registration.Bank == "" {
			return hindsightConfiguration{}, errors.New("chaque inscription exige repository absolu et bank")
		}
		repository, err := expandHindsightRepository(registration.Repository)
		if err != nil {
			return hindsightConfiguration{}, err
		}
		registration.Repository = repository
		info, err := os.Stat(repository)
		if err != nil || !info.IsDir() {
			return hindsightConfiguration{}, fmt.Errorf("dépôt introuvable : %s", repository)
		}
		if seenRepositories[repository] {
			return hindsightConfiguration{}, fmt.Errorf("dépôt enregistré deux fois : %s", repository)
		}
		seenRepositories[repository] = true
	}
	return configuration, nil
}

func expandHindsightRepository(repository string) (string, error) {
	if filepath.IsAbs(repository) {
		return repository, nil
	}
	if repository == "~" || strings.HasPrefix(repository, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, strings.TrimPrefix(repository, "~/")), nil
	}
	return "", errors.New("chaque inscription exige repository absolu ou commençant par ~/")
}

func registerHindsightFiles(home string, configuration hindsightConfiguration) error {
	if home == "" {
		return errors.New("HOME est absent")
	}
	managedPath := filepath.Join(home, ".hindsight", "dotfiles-managed.json")
	managed, err := readHindsightManaged(managedPath)
	if err != nil {
		return err
	}
	agentPath := filepath.Join(home, ".hindsight", "coding-agent.json")
	cursorPath := filepath.Join(home, ".cursor", "mcp.json")
	if err := validateHindsightJSON(agentPath, func(document map[string]any) error { return updateHindsightAgent(document, configuration, managed) }); err != nil {
		return err
	}
	if err := validateHindsightJSON(cursorPath, func(document map[string]any) error { return updateHindsightCursor(document, configuration, managed) }); err != nil {
		return err
	}
	if err := mergeHindsightAgentConfig(agentPath, configuration, managed); err != nil {
		return err
	}
	if err := mergeHindsightCursorConfig(cursorPath, configuration, managed); err != nil {
		return err
	}
	return writeHindsightManaged(managedPath, configuration)
}

func mergeHindsightAgentConfig(path string, configuration hindsightConfiguration, managed hindsightManaged) error {
	return mergeHindsightJSON(path, func(document map[string]any) error { return updateHindsightAgent(document, configuration, managed) })
}

func updateHindsightAgent(document map[string]any, configuration hindsightConfiguration, managed hindsightManaged) error {
	document["apiUrl"] = configuration.APIURL
	document["apiToken"] = configuration.APIToken
	document["optInOnly"] = true
	document["autoUpdate"] = false
	paths, err := objectField(document, "mapPathToBank")
	if err != nil {
		return err
	}
	for _, repository := range managed.Repositories {
		delete(paths, repository)
	}
	for _, registration := range configuration.Registrations {
		paths[registration.Repository] = registration.Bank
	}
	return nil
}

func mergeHindsightCursorConfig(path string, configuration hindsightConfiguration, managed hindsightManaged) error {
	return mergeHindsightJSON(path, func(document map[string]any) error { return updateHindsightCursor(document, configuration, managed) })
}

func updateHindsightCursor(document map[string]any, configuration hindsightConfiguration, managed hindsightManaged) error {
	servers, err := objectField(document, "mcpServers")
	if err != nil {
		return err
	}
	for _, bank := range managed.Banks {
		delete(servers, hindsightServerName(bank))
	}
	seenNames := map[string]string{}
	for _, registration := range configuration.Registrations {
		name := hindsightServerName(registration.Bank)
		if name == "hindsight-memory-" {
			return errors.New("nom de banque incompatible avec Cursor")
		}
		if bank, exists := seenNames[name]; exists && bank != registration.Bank {
			return fmt.Errorf("banques Cursor ambiguës : %s et %s", bank, registration.Bank)
		}
		seenNames[name] = registration.Bank
		url := configuration.APIURL + "/mcp/" + url.PathEscape(registration.Bank) + "/"
		if current, exists := servers[name]; exists {
			server, ok := current.(map[string]any)
			if !ok || server["url"] != url {
				return fmt.Errorf("mcpServers.%s existe déjà et ne cible pas cette banque", name)
			}
		}
		servers[name] = map[string]any{
			"type":    "http",
			"url":     url,
			"headers": map[string]any{"Authorization": "Bearer " + configuration.APIToken},
		}
	}
	return nil
}

func hindsightServerName(bank string) string {
	name := strings.Trim(nonServerName.ReplaceAllString(strings.ToLower(bank), "-"), "-")
	return "hindsight-memory-" + name
}

func objectField(document map[string]any, name string) (map[string]any, error) {
	value, exists := document[name]
	if !exists {
		object := map[string]any{}
		document[name] = object
		return object, nil
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("champ %s incompatible", name)
	}
	return object, nil
}

func mergeHindsightJSON(path string, update func(map[string]any) error) error {
	content, _, existed, err := readSettings(path)
	if err != nil {
		return err
	}
	var document map[string]any
	if err := json.Unmarshal(content, &document); err != nil {
		return errors.New("JSON invalide")
	}
	if document == nil {
		return errors.New("JSON doit être un objet")
	}
	if err := update(document); err != nil {
		return err
	}
	return replaceJSON(path, content, document, 0o600, existed)
}

func validateHindsightJSON(path string, update func(map[string]any) error) error {
	content, _, _, err := readSettings(path)
	if err != nil {
		return err
	}
	var document map[string]any
	if err := json.Unmarshal(content, &document); err != nil || document == nil {
		return errors.New("JSON invalide")
	}
	return update(document)
}

func readHindsightManaged(path string) (hindsightManaged, error) {
	content, _, _, err := readSettings(path)
	if err != nil {
		return hindsightManaged{}, err
	}
	var managed hindsightManaged
	if err := json.Unmarshal(content, &managed); err != nil {
		return hindsightManaged{}, errors.New("état Hindsight invalide")
	}
	return managed, nil
}

func writeHindsightManaged(path string, configuration hindsightConfiguration) error {
	content, _, existed, err := readSettings(path)
	if err != nil {
		return err
	}
	banks := map[string]bool{}
	managed := hindsightManaged{}
	for _, registration := range configuration.Registrations {
		managed.Repositories = append(managed.Repositories, registration.Repository)
		if !banks[registration.Bank] {
			managed.Banks = append(managed.Banks, registration.Bank)
			banks[registration.Bank] = true
		}
	}
	document := map[string]any{"repositories": managed.Repositories, "banks": managed.Banks}
	return replaceJSON(path, content, document, 0o600, existed)
}
