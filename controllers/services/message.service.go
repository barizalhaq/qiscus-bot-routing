package services

import (
	"bot-routing-engine/entities/viewmodel"
	"bot-routing-engine/repositories"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"time"
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

	if !s.isOnWorkingHour() {
		notInWorkingHourDraft := viewmodel.Draft{
			Room: input,
			Layer: viewmodel.Layer{
				Resolve: true,
			},
			Message: os.Getenv("NOT_IN_WORKING_HOUR_WORDING"),
		}
		drafts = append(drafts, notInWorkingHourDraft)

		return drafts, nil
	}

	var states []int
	option := input.Payload.Message.Text
	if !s.room.StateExist(roomInfo) {
		s.room.UpdateBotState(input.Payload.Room.ID, states, roomInfo)
		draft.Message = layer.Message
		drafts = append(drafts, draft)
	} else {
		var jsonOptions map[string]json.RawMessage
		json.Unmarshal([]byte(roomInfo.Results.Rooms[0].Options), &jsonOptions)
		json.Unmarshal(jsonOptions["bot_layer"], &states)

		if directKeypad, directAssignEnable := os.LookupEnv("DIRECT_ASSIGN_AGENT_KEYPAD"); directAssignEnable && input.Payload.Message.Text == directKeypad && len(states) == 0 {
			directHandoverDraft := viewmodel.Draft{
				Room: input,
				Layer: viewmodel.Layer{
					Handover: true,
				},
				Message: os.Getenv("WAITING_FOR_AGENT_WORDING"),
			}
			drafts = append(drafts, directHandoverDraft)
			return
		}

		choosenLayer, err := NewLayerService().DetermineLayer(option, states, layer)
		if err != nil {
			draft.Message = err.Error()
			drafts = append(drafts, draft)
			draft.Message = choosenLayer.Message
			drafts = append(drafts, draft)
			return drafts, nil
		}

		if prevLayerKeypad, prevLayerKeypadEnable := os.LookupEnv("RETURN_PREVIOUS_LAYER_KEYPAD"); prevLayerKeypadEnable &&
			input.Payload.Message.Text == prevLayerKeypad && len(states) > 0 {
			states = states[:len(states)-1]
		} else {
			optionInt, _ := strconv.Atoi(option)
			states = append(states, optionInt)
		}
		s.room.UpdateBotState(input.Payload.Room.ID, states, roomInfo)
		draft.Message = choosenLayer.Message
		draft.Layer = choosenLayer
		drafts = append(drafts, draft)
		return drafts, nil
	}

	return
}

func (s *messageService) isOnWorkingHour() bool {
	officeHour, _ := s.multichannelRepository.OfficeHour()
	loc, _ := time.LoadLocation(os.Getenv("TIMEZONE"))

	now := time.Now().In(loc)
	for _, day := range officeHour.Data.OfficeHours {
		if int(now.Weekday()) == day.Day || int(time.Saturday)+1 == day.Day {
			officeHourStartTime := fmt.Sprintf("%d-%d-%d %s", now.Year(), now.Month(), now.Day(), day.Starttime)
			officeHourEndTime := fmt.Sprintf("%d-%d-%d %s", now.Year(), now.Month(), now.Day(), day.Endtime)

			layout := "2006-1-2 15:04"
			parsedStartTime, err := time.Parse(layout, officeHourStartTime)
			if err != nil {
				panic(err.Error())
			}
			parsedEndTime, err := time.Parse(layout, officeHourEndTime)
			if err != nil {
				panic(err.Error())
			}

			if now.Hour() >= parsedStartTime.Hour() && now.Local().Minute() >= parsedStartTime.Minute() {
				if now.Hour() <= parsedEndTime.Hour() {
					return true
				} else if now.Hour() == parsedEndTime.Hour() {
					return now.Local().Minute() <= parsedEndTime.Minute()
				}
			}

			return false
		}
	}

	return false
}
