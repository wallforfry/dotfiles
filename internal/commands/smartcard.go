package commands

import (
	"encoding/json"
	"io"
	"strings"
)

// Une élévation temporaire de privilèges relance com.apple.ctkpcscd : le
// contexte PC/SC détenu par scdaemon meurt et scdaemon ne le rétablit jamais,
// jusqu'à ce qu'on le tue. Signatures observées côté appelant, gpg étant
// localisé.
var smartcardFailures = []string{
	"card not available",
	"carte non disponible",
	"card error",
	"erreur de carte",
	"card removed",
	"carte retirée",
	"selecting card failed",
	"échec de la sélection de la carte",
	"no such device",
	"aucun périphérique de ce type",
	"agent refused operation",
	"gpg failed to sign the data",
}

const smartcardHookPayloadLimit = 1 << 20

func SmartcardWakeup(runtime Runtime, args []string) int {
	switch firstArg(args) {
	case "":
		return wakeSmartcard(runtime)
	case "--hook":
		return smartcardHook(runtime)
	default:
		fprintf(runtime.Stderr, "smartcard-wakeup: usage - sans argument, ou --hook\n")
		return ExitUsage
	}
}

func wakeSmartcard(runtime Runtime) int {
	if smartcardReachable(runtime) {
		fprintf(runtime.Stdout, "🔑  carte déjà disponible\n")
		return 0
	}
	if repairSmartcard(runtime) {
		fprintf(runtime.Stdout, "🔑  carte réveillée : scdaemon redémarré\n")
		return 0
	}
	fprintf(runtime.Stderr, "🔑  carte toujours indisponible après redémarrage de scdaemon\n")
	return ExitUnavailable
}

// Le hook ne parle que lorsqu'il a réparé quelque chose : une machine sans carte
// - Linux, DSM - échouerait sinon à chaque signature reconnue, sans recours.
func smartcardHook(runtime Runtime) int {
	if !mentionsSmartcardFailure(hookToolResponse(runtime.Stdin)) {
		return 0
	}
	if smartcardReachable(runtime) || !repairSmartcard(runtime) {
		return 0
	}
	fprintf(runtime.Stderr, "🔑  carte réveillée : scdaemon redémarré, relance la commande.\n")
	return ExitFeedback
}

// Seule la réponse de l'outil est examinée : la commande soumise contient les
// signatures dès qu'on lit ou teste ce fichier.
func hookToolResponse(stdin io.Reader) string {
	payload, err := io.ReadAll(io.LimitReader(stdin, smartcardHookPayloadLimit))
	if err != nil {
		return ""
	}
	var event struct {
		ToolResponse json.RawMessage `json:"tool_response"`
	}
	if json.Unmarshal(payload, &event) != nil {
		return ""
	}
	return string(event.ToolResponse)
}

func mentionsSmartcardFailure(response string) bool {
	lowered := strings.ToLower(response)
	for _, failure := range smartcardFailures {
		if strings.Contains(lowered, failure) {
			return true
		}
	}
	return false
}

func smartcardReachable(runtime Runtime) bool {
	return runtime.Executor.Run(runtime.silent("gpg", "--card-status")) == nil
}

func repairSmartcard(runtime Runtime) bool {
	_ = runtime.Executor.Run(runtime.silent("gpgconf", "--kill", "scdaemon"))
	return smartcardReachable(runtime)
}
