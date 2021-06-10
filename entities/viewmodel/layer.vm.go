package viewmodel

type Layer struct {
	Message  string  `json:"message"`
	Options  []Layer `json:"options"`
	Handover bool    `json:"handover"`
	Input    bool    `json:"input"`
	Resolve  bool    `json:"resolve"`
	Division string  `json:"division"`
}
