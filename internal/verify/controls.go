package verify

type Control string

const (
	ControlSyntax      Control = "syntax"
	ControlGo          Control = "go"
	ControlTemplates   Control = "templates"
	ControlBootstrap   Control = "bootstrap"
	ControlSkills      Control = "skills"
	ControlRouting     Control = "routing"
	ControlSubagents   Control = "subagents"
	ControlProjections Control = "projections"
	ControlWorkflows   Control = "workflows"
	ControlADR         Control = "adr"
	ControlSensitive   Control = "sensitive"
	ControlEncryption  Control = "encryption"
	ControlLiveState   Control = "live-state"
	ControlTelemetry   Control = "telemetry"
)

func allControls() []Control {
	return []Control{
		ControlSyntax, ControlGo, ControlTemplates, ControlBootstrap, ControlSkills, ControlRouting,
		ControlSubagents, ControlProjections, ControlWorkflows, ControlADR,
		ControlSensitive, ControlEncryption, ControlLiveState, ControlTelemetry,
	}
}

func (v *verifier) check(control Control) (func(), bool) {
	checks := map[Control]func(){
		ControlSyntax: v.checkSyntax, ControlGo: v.checkGo, ControlTemplates: v.checkTemplates,
		ControlBootstrap: v.checkBootstrap,
		ControlSkills:    v.checkSkills, ControlRouting: v.checkRouting, ControlSubagents: v.checkSubagents,
		ControlProjections: v.checkProjections, ControlWorkflows: v.checkWorkflows, ControlADR: v.checkADR,
		ControlSensitive: v.checkSensitive, ControlEncryption: v.checkEncryption,
		ControlLiveState: v.checkLiveState, ControlTelemetry: v.checkTelemetry,
	}
	check, ok := checks[control]
	return check, ok
}
