package message

import (
	"bot-routing-engine/entities/viewmodel"
	"fmt"
)

func FormConfirmationMessage(userInfo []map[string]string, layer viewmodel.Layer) string {
	messageLayout := layer.AdditionalInformation.FormsConfirmation.Message

	var dataAsMessage string
	for _, form := range layer.AdditionalInformation.Forms {
		for _, info := range userInfo {
			if form.Key == info["key"] {
				var list string
				if len(form.EngKey) > 0 {
					list = fmt.Sprintf("%s: %s\n%s: %s\n\n", info["key"], info["value"], form.EngKey, info["value"])
				} else {
					list = fmt.Sprintf("%s: %s\n", info["key"], info["value"])
				}
				dataAsMessage = dataAsMessage + list
			}
		}
	}

	confirmationMessage := fmt.Sprintf(messageLayout, dataAsMessage)

	return confirmationMessage
}
