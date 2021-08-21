package viewmodel

type Layer struct {
	Message                  string                    `json:"message"`
	Options                  []Layer                   `json:"options"`
	Handover                 bool                      `json:"handover"`
	Input                    bool                      `json:"input"`
	Resolve                  bool                      `json:"resolve"`
	Division                 string                    `json:"division"`
	AddAdditionalInformation bool                      `json:"add_additional_information"`
	AdditionalInformation    AdditionalInformationType `json:"additional_information"`
}

type AdditionalInformationType struct {
	Forms []AdditionalInformationForm `json:"forms"`
}

type AdditionalInformationForm struct {
	Question string `json:"question"`
	Key      string `json:"key"`
}

type AdditionalInformationConfirmationOption struct {
	Begin   bool `json:"begin"`
	Resolve bool `json:"resolve"`
}
