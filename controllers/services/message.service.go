package services

import (
	"bot-routing-engine/entities/viewmodel"
	"bot-routing-engine/repositories"
	"bot-routing-engine/utils/message"
	"encoding/json"
	"fmt"
	"log"
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

			if len(choosenLayer.Messages) > 0 {
				for _, msg := range choosenLayer.Messages {
					draft.Message = msg
					drafts = append(drafts, draft)
				}
			} else {
				draft.Message = choosenLayer.Message
				drafts = append(drafts, draft)
			}
			return drafts, nil
		}

		if prevLayerKeypad, prevLayerKeypadEnable := os.LookupEnv("RETURN_PREVIOUS_LAYER_KEYPAD"); prevLayerKeypadEnable &&
			input.Payload.Message.Text == prevLayerKeypad && len(states) > 0 {
			states = states[:len(states)-1]
		} else {
			optionInt, _ := strconv.Atoi(option)
			states = append(states, optionInt)
		}

		var formState int
		json.Unmarshal(jsonOptions["forms_layer_index"], &formState)

		if jsonOptions["forms_layer_index"] == nil {
			s.room.UpdateBotState(input.Payload.Room.ID, states, roomInfo)
		}

		if choosenLayer.AddAdditionalInformation {
			drafts = s.handleAdditionalInformation(input.Payload.Room.ID, drafts, option, choosenLayer, formState, input)
		} else {
			draft.Layer = choosenLayer
			if len(choosenLayer.Messages) > 0 {
				for _, msg := range choosenLayer.Messages {
					draft.Message = msg
					drafts = append(drafts, draft)
				}
			} else {
				draft.Message = choosenLayer.Message
				drafts = append(drafts, draft)
			}
		}

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

			if now.Hour() > parsedStartTime.Hour() {
				if now.Hour() < parsedEndTime.Hour() {
					return true
				} else if now.Hour() == parsedEndTime.Hour() {
					return now.Minute() <= parsedEndTime.Minute()
				}

				return false
			} else if now.Hour() == parsedStartTime.Hour() {
				if now.Minute() >= parsedStartTime.Minute() {
					if now.Hour() < parsedEndTime.Hour() {
						return true
					} else if now.Hour() == parsedEndTime.Hour() {
						return now.Minute() <= parsedEndTime.Minute()
					}
				}

				return false
			}

			return false
		}
	}

	return false
}

func (s *messageService) handleAdditionalInformation(roomID string, existingDrafts []viewmodel.Draft, textMessage string, layer viewmodel.Layer, latestFormState int, input *viewmodel.WebhookRequest) []viewmodel.Draft {
	roomInfo, err := s.room.SDKGetRoomInfo(input.Payload.Room.ID)
	if err != nil {
		log.Fatalf("Something went wrong: %s", err.Error())
		return []viewmodel.Draft{}
	}
	formStateExist := s.room.roomRepository.FormStateExist(roomInfo)
	draft := viewmodel.Draft{
		Room:  input,
		Layer: layer,
	}

	if !formStateExist {
		draft.Message = os.Getenv("ADDITIONAL_INFORMATION_INSTRUCTION")
		existingDrafts = append(existingDrafts, draft)
		draft.Message = layer.AdditionalInformation.Forms[0].Question
		existingDrafts = append(existingDrafts, draft)
		s.room.UpdateFormsState(roomID, 0, roomInfo)
		return existingDrafts
	}

	newFormData := map[string]string{
		"key":   layer.AdditionalInformation.Forms[latestFormState].Key,
		"value": textMessage,
	}
	s.room.SaveNewFormData(roomID, newFormData)

	nextFormIndex := latestFormState + 1
	// Form is over
	/*
		if nextFormIndex >= len(layer.AdditionalInformation.Forms) {
			userInfo, _ := s.room.roomRepository.GetRoomUserInfo(roomID)
			confirmationMessage := message.FormConfirmationMessage(userInfo.Data.Extras.UserProperties, layer)
			draft.Message = confirmationMessage
			formConfirmed := s.room.roomRepository.FormConfirmed(roomInfo)

			if !s.room.roomRepository.FormConfirmedExist(roomInfo) {
				existingDrafts = append(existingDrafts, draft)
				s.room.SetFormConfirmationStatus(roomID, false, roomInfo)
			} else if !formConfirmed {
				confirmed, msg, err := NewLayerService().GetFormConfirmationOption(textMessage, layer)
				if err != nil {
					draft.Message = err.Error()
					existingDrafts = append(existingDrafts, draft)
					draft.Message = confirmationMessage
					existingDrafts = append(existingDrafts, draft)
				} else if !confirmed && err == nil {
					draft.Message = msg
					existingDrafts = append(existingDrafts, draft)

					s.room.SetFormConfirmationStatus(roomID, false, roomInfo)
				} else {
					formConfirmedResp := viewmodel.Draft{
						Message: msg,
						Layer: viewmodel.Layer{
							Handover: true,
						},
						Room: input,
					}
					s.room.SetFormConfirmationStatus(roomID, true, roomInfo)
					existingDrafts = append(existingDrafts, formConfirmedResp)
				}
			}
			return existingDrafts
		}
	*/

	if nextFormIndex >= len(layer.AdditionalInformation.Forms) {
		userInfo, _ := s.room.roomRepository.GetRoomUserInfo(roomID)
		confirmationMessage := message.FormConfirmationMessage(userInfo.Data.Extras.UserProperties, layer)

		if len(layer.AdditionalInformation.FormsConfirmation.AdditionalMessages) > 0 {
			for _, msg := range layer.AdditionalInformation.FormsConfirmation.AdditionalMessages {
				draft.Message = msg
				existingDrafts = append(existingDrafts, draft)
			}
		}

		formOverDraftMessage := viewmodel.Draft{
			Message: confirmationMessage,
			Layer: viewmodel.Layer{
				Handover: true,
				Division: layer.Division,
			},
			Room: input,
		}

		existingDrafts = append(existingDrafts, formOverDraftMessage)
		return existingDrafts
	}

	nextForm := layer.AdditionalInformation.Forms[nextFormIndex]
	draft.Message = nextForm.Question

	existingDrafts = append(existingDrafts, draft)
	s.room.UpdateFormsState(roomID, nextFormIndex, roomInfo)
	return existingDrafts
}
