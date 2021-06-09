package services

import (
	"bot-routing-engine/entities/viewmodel"
	"bot-routing-engine/repositories"
	"encoding/json"
	"os"
	"strconv"
)

type MessageService interface {
	Determine(request interface{}) (drafts []viewmodel.Draft, err error)
}

type messageService struct {
	multichannelRepository repositories.MultichannelRepository
	room                   roomService
}

func NewMessageService(multichannelRepository repositories.MultichannelRepository, room roomService) *messageService {
	return &messageService{multichannelRepository, room}
}

func (s *messageService) Determine(request interface{}) (drafts []viewmodel.Draft, err error) {
	input := request.(*viewmodel.WebhookRequest)

	// multichannel working hour

	var roomOption viewmodel.Option
	json.Unmarshal([]byte(input.Payload.Room.Options), &roomOption)

	roomInfo, err := s.room.SDKGetRoomInfo(input.Payload.Room.ID)
	if err != nil {
		return
	}

	layer, err := NewLayerService().GetLayer(roomOption.ChannelDetails.ChannelID)
	if err != nil {
		return
	}

	draft := viewmodel.Draft{
		Room:  input,
		Layer: layer,
	}

	var states []int
	option := input.Payload.Message.Text
	if !s.room.StateExist(roomInfo) {
		s.room.UpdateBotState(input.Payload.Room.ID, states, roomInfo)
		draft.Message = layer.Message
		drafts = append(drafts, draft)
	} else {
		if directKeypad, directAssignEnable := os.LookupEnv("DIRECT_ASSIGN_AGENT_KEYPAD"); directAssignEnable && input.Payload.Message.Text == directKeypad {
			directHandoverDraft := &draft
			directHandoverDraft.Layer = viewmodel.Layer{
				Message:  "Mohon tunggu sebentar, Anda akan terhubung dengan agent Brodo",
				Handover: true,
			}
			drafts = append(drafts, draft)
			return
		}

		var jsonOptions map[string]json.RawMessage
		json.Unmarshal([]byte(roomInfo.Results.Rooms[0].Options), &jsonOptions)
		json.Unmarshal(jsonOptions["bot_layer"], &states)

		choosenLayer, err := NewLayerService().DetermineLayer(option, states, layer)
		if err != nil {
			draft.Message = err.Error()
			drafts = append(drafts, draft)
			draft.Message = choosenLayer.Message
			drafts = append(drafts, draft)
			return drafts, nil
		}

		optionInt, _ := strconv.Atoi(option)
		states = append(states, optionInt)

		s.room.UpdateBotState(input.Payload.Room.ID, states, roomInfo)
		draft.Message = choosenLayer.Message
		draft.Layer = choosenLayer
		drafts = append(drafts, draft)
	}

	return
}
