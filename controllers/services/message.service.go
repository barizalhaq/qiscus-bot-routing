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
		s.room.UpdateBotState(input.Payload.Room.ID, states)
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

		choosenLayer, err := NewLayerService().DetermineLayer(option, states, layer, jsonOptions)
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

		var formState int
		json.Unmarshal(jsonOptions["forms_layer_index"], &formState)

		if prevLayerKeypad, prevLayerKeypadEnable := os.LookupEnv("RETURN_PREVIOUS_LAYER_KEYPAD"); prevLayerKeypadEnable &&
			input.Payload.Message.Text == prevLayerKeypad && len(states) > 0 && jsonOptions["forms_layer_index"] == nil {
			states = states[:len(states)-1]
		} else if resetLayerKeypad, resetLayerEnable := os.LookupEnv("RESET_LAYER_KEYPAD"); resetLayerEnable &&
			input.Payload.Message.Text == resetLayerKeypad && len(states) > 0 && jsonOptions["forms_layer_index"] == nil {
			lastIndex, _ := strconv.Atoi(os.Getenv("RESET_LAST_INDEX"))
			states = states[:lastIndex]
		} else {
			optionInt, _ := strconv.Atoi(option)
			states = append(states, optionInt)
		}

		if jsonOptions["forms_layer_index"] == nil {
			s.room.UpdateBotState(input.Payload.Room.ID, states)
		}

		if choosenLayer.AddAdditionalInformation {
			layers := map[string]viewmodel.Layer{
				"existing": choosenLayer,
				"channel":  layer,
			}
			drafts = s.handleAdditionalInformation(input.Payload.Room.ID, drafts, option, layers, formState, input, jsonOptions)
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

func (s *messageService) handleAdditionalInformation(roomID string, existingDrafts []viewmodel.Draft, textMessage string, layer map[string]viewmodel.Layer, latestFormState int, input *viewmodel.WebhookRequest, jsonOptions map[string]json.RawMessage) []viewmodel.Draft {
	state := textMessage

	roomInfo, err := s.room.SDKGetRoomInfo(input.Payload.Room.ID)
	if err != nil {
		log.Fatalf("Something went wrong: %s", err.Error())
		return []viewmodel.Draft{}
	}
	formStateExist := s.room.roomRepository.FormStateExist(roomInfo)
	draft := viewmodel.Draft{
		Room:  input,
		Layer: layer["existing"],
	}

	if !formStateExist {
		draft.Message = layer["existing"].AdditionalInformation.Instruction
		existingDrafts = append(existingDrafts, draft)
		draft.Message = layer["existing"].AdditionalInformation.Forms[0].Question
		existingDrafts = append(existingDrafts, draft)
		s.room.UpdateFormsState(roomID, 0)
		return existingDrafts
	}

	formIsConfirming := s.room.FormConfirming(roomID)
	/*
		If multiple answers required
	*/
	ongoingForm := layer["existing"].AdditionalInformation.Forms[latestFormState]
	var multipleAnswerState int
	if !formIsConfirming {
		if len(ongoingForm.Answers) > 0 {
			textMessage, multipleAnswerState, err = NewLayerService().GetAnswer(textMessage, ongoingForm)
			if err != nil {
				draft.Message = err.Error()
				existingDrafts = append(existingDrafts, draft)

				draft.Message = ongoingForm.Question
				existingDrafts = append(existingDrafts, draft)
				return existingDrafts
			}
		}

		/*
			Nested question
		*/
		if len(ongoingForm.Questions) > 0 {
			key := fmt.Sprintf("%s_question_index", ongoingForm.Key)

			var nestedQuestionKeys map[string]int
			json.Unmarshal(jsonOptions["nested_form_keys"], &nestedQuestionKeys)

			nestedQuestion := ongoingForm.Questions[nestedQuestionKeys[key]]

			/*
				Multiple answers in nested question
			*/
			if len(nestedQuestion.Answers) > 0 {
				textMessage, err = NewLayerService().GetNestedQuestionAnswer(textMessage, nestedQuestion)
				if err != nil {
					draft.Message = err.Error()
					existingDrafts = append(existingDrafts, draft)

					draft.Message = nestedQuestion.Question
					existingDrafts = append(existingDrafts, draft)
					return existingDrafts
				}
			}
		}
		/*
			Save to room additional information first
		*/
		newFormData := map[string]string{
			"key":   layer["existing"].AdditionalInformation.Forms[latestFormState].Key,
			"value": textMessage,
		}
		s.room.SaveNewFormData(roomID, newFormData)
	}

	/*
		If forms is over asking question
	*/
	nextFormIndex := latestFormState + 1
	if nextFormIndex >= len(layer["existing"].AdditionalInformation.Forms) {
		userInfo, _ := s.room.roomRepository.GetRoomUserInfo(roomID)
		confirmationMessage := message.FormConfirmationMessage(userInfo.Data.Extras.UserProperties, layer["existing"])

		if formIsConfirming {
			resetLayerKeypad, resetLayerEnable := os.LookupEnv("RESET_LAYER_KEYPAD")
			if resetLayerEnable && state == resetLayerKeypad {
				s.room.roomRepository.DeleteRoomOption(roomID, "forms_layer_index")
				s.room.roomRepository.DeleteRoomOption(roomID, "nested_form_keys")
				s.room.roomRepository.DeleteRoomOption(roomID, "form_confirming")

				states := s.room.ResetLayer(roomID)

				choosenLayer := NewLayerService().getLatestLayer(states, layer["channel"])
				draft.Message = choosenLayer.Message

				existingDrafts = append(existingDrafts, draft)

				return existingDrafts
			}

			option, err := NewLayerService().FormConfirmationOption(state, layer["existing"].AdditionalInformation.FormsConfirmation.Options)
			if err != nil {
				draft.Message = err.Error()
				existingDrafts = append(existingDrafts, draft)

				draft.Message = confirmationMessage
				existingDrafts = append(existingDrafts, draft)

				return existingDrafts
			}

			if option.Confirmed {
				additionalMessages := layer["existing"].AdditionalInformation.FormsConfirmation.AdditionalMessages
				if len(additionalMessages) > 0 {
					for i, msg := range additionalMessages {
						if i == len(additionalMessages)-1 {
							draft.Layer = viewmodel.Layer{
								Handover: true,
								Division: layer["existing"].Division,
							}
						}
						draft.Message = msg
						existingDrafts = append(existingDrafts, draft)
					}
				}

				return existingDrafts
			}

			if option.Reset {
				s.room.roomRepository.DeleteRoomOption(roomID, "forms_layer_index")
				s.room.roomRepository.DeleteRoomOption(roomID, "nested_form_keys")
				s.room.roomRepository.DeleteRoomOption(roomID, "form_confirming")

				draft.Message = layer["existing"].AdditionalInformation.Instruction
				existingDrafts = append(existingDrafts, draft)
				draft.Message = layer["existing"].AdditionalInformation.Forms[0].Question
				existingDrafts = append(existingDrafts, draft)
				s.room.UpdateFormsState(roomID, 0)
				return existingDrafts
			}
		}

		draft.Message = confirmationMessage
		existingDrafts = append(existingDrafts, draft)

		s.room.SetFormConfirming(roomID, true)
		return existingDrafts
	}

	/*
		Customer get next form question
	*/
	nextForm := layer["existing"].AdditionalInformation.Forms[nextFormIndex]
	if len(nextForm.Questions) > 0 {
		for _, index := range ongoingForm.RequiredBy {
			key := fmt.Sprintf("%s_question_index", layer["existing"].AdditionalInformation.Forms[index].Key)
			s.room.UpdateNestedFormState(roomID, multipleAnswerState, key)
		}
		draft.Message = nextForm.Questions[multipleAnswerState].Question
	} else {
		draft.Message = nextForm.Question
	}
	existingDrafts = append(existingDrafts, draft)

	s.room.UpdateFormsState(roomID, nextFormIndex)
	return existingDrafts
}
