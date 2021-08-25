package message

import (
	"bot-routing-engine/entities/viewmodel"
	"fmt"
)

func FormConfirmationMessage(userInfo []map[string]string, layer viewmodel.Layer) string {
	messageLayout := layer.AdditionalInformation.FormsConfirmation.Message

	var dataAsMessage string
	for _, info := range userInfo {
		list := fmt.Sprintf("%s: %s\n", info["key"], info["value"])
		dataAsMessage = dataAsMessage + list
	}

	confirmationMessage := fmt.Sprintf(messageLayout, dataAsMessage)

	return confirmationMessage
}
