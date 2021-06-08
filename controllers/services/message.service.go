package services

import (
	"bot-routing-engine/entities"
	"bot-routing-engine/entities/viewmodel"
	"bot-routing-engine/repositories"
	"encoding/json"
)

type MessageService interface {
	Determine(request interface{}) (layer viewmodel.Layer, roomInfo entities.Room, err error)
	SendBotMessage(roomID string, message string) error
}

type messageService struct {
	roomRepository repositories.RoomRepository
	multichannelRepository repositories.MultichannelRepository
}

func NewMessageService(roomRepo repositories.RoomRepository, multichannelRepository repositories.MultichannelRepository) *messageService {
	return &messageService{roomRepo, multichannelRepository}
}

func (s *messageService) Determine(request interface{}) (layer viewmodel.Layer, roomInfo entities.Room, err error) {
	input := request.(*viewmodel.WebhookRequest)

	// multichannel working hour

	var roomOption viewmodel.Option
	json.Unmarshal([]byte(input.Payload.Room.Options), &roomOption)

	roomInfo, err = s.roomRepository.GetRoomInfo(input.Payload.Room.ID)
	if err != nil {
		return
	}

	layer, err = NewLayerService().GetLayer(roomOption.ChannelDetails.ChannelID)
	if err != nil {
		return
	}
	// if s.roomRepository.StateExist(roomInfo) {
		
	// }

	return
}

func (s *messageService) SendBotMessage(roomID string, message string) error {
	err := s.multichannelRepository.SendBotMessage(roomID, message)

	if err != nil {
		return err
	}

	return nil
}