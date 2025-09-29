package model

type DetectionSummary struct {
	Filename       string   `json:"filename,omitempty"`
	Classification string   `json:"classification,omitempty"`
	RiskLevel      string   `json:"risk_level,omitempty"`
	Confidence     string   `json:"confidence,omitempty"`
	ActionRequired string   `json:"action_required,omitempty"`
	Summary        string   `json:"summary,omitempty"`
	KeyFindings    []string `json:"key_findings,omitempty"`
	ProcessingTime string   `json:"processing_time,omitempty"`
}
