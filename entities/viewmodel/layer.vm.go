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
	Messages                 []string                  `json:"messages"`
}

type AdditionalInformationType struct {
	Forms             []AdditionalInformationForm `json:"forms"`
	FormsConfirmation struct {
		Message            string                    `json:"message"`
		Options            []FormsConfirmationOption `json:"options"`
		AdditionalMessages []string                  `json:"additional_messages"`
	} `json:"forms_confirmation"`
}

type AdditionalInformationForm struct {
	Question string `json:"question"`
	Key      string `json:"key"`
	EngKey   string `json:"eng_key"`
}

type AdditionalInformationConfirmationOption struct {
	Begin   bool `json:"begin"`
	Resolve bool `json:"resolve"`
}

type FormsConfirmationOption struct {
	Confirmed bool   `json:"confirmed"`
	Message   string `json:"message"`
}
