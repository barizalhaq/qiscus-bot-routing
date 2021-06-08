package viewmodel

type Layer struct {
	Message  string                 `json:"message"`
	Options  map[string]interface{} `json:"options"`
	Handover bool                   `json:"handover"`
	Input    bool                   `json:"input"`
	Resolve  bool                   `json:"resolve"`
}
