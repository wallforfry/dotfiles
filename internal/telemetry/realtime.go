package telemetry

import "slices"

var codexRealtimePayloadTypes = []string{
	"realtime_session_closed", "realtime_session_started", "transcript_segment",
}

func validRealtimePayload(payload map[string]any) bool {
	payloadType, ok := payload["type"].(string)
	if !ok || !slices.Contains(codexRealtimePayloadTypes, payloadType) ||
		!stringFields(payload, "id", "realtime_session_id") {
		return false
	}
	switch payloadType {
	case "transcript_segment":
		role, roleOK := payload["role"].(string)
		_, textOK := payload["text"].(string)
		return roleOK && slices.Contains([]string{"assistant", "user"}, role) && textOK
	case "realtime_session_closed":
		return stringFields(payload, "outcome")
	default:
		return true
	}
}

func integerFields(object map[string]any, keys ...string) bool {
	for _, key := range keys {
		if !integerValue(object[key]) {
			return false
		}
	}
	return true
}

func integerValue(value any) bool {
	switch typed := value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	case float64:
		return typed == float64(int64(typed))
	default:
		return false
	}
}
