package services

import (
	"bot-routing-engine/entities"
	"bot-routing-engine/entities/viewmodel"
	"bot-routing-engine/repositories"
	"bot-routing-engine/utils/agent"
	"encoding/json"
	"os"
	"strconv"
)

type RoomService interface {
	SendBotMessage(roomID string, message string) error
	Resolve(roomID string, lastCommentID string) error
	SDKGetRoomInfo(ID string) (entities.Room, error)
	UpdateBotState(roomID string, states []int, roomInfo entities.Room) error
	StateExist(room entities.Room) bool
	QismoRoomInfo(ID string) (viewmodel.QismoRoomInfo, error)
	AutoResolveTag(ID string) error
	AutoHandover(ID string) error
	Handover(ID string) error
}

type roomService struct {
	multichannelRepository repositories.MultichannelRepository
	roomRepository         repositories.RoomRepository
}

func NewRoomService(multichannelRepository repositories.MultichannelRepository, roomRepository repositories.RoomRepository) *roomService {
	return &roomService{multichannelRepository, roomRepository}
}

func (s *roomService) SendBotMessage(roomID string, message string) error {
	err := s.multichannelRepository.SendBotMessage(roomID, message)

	if err != nil {
		return err
	}

	return nil
}

func (s *roomService) Resolve(roomID string, lastCommentID string) error {
	err := s.roomRepository.ResetBotLayers(roomID)
	if err != nil {
		return err
	}

	if os.Getenv("ENABLE_AUTO_RESOLVE_TAG") == "true" {
		err = s.AutoResolveTag(roomID)
		if err != nil {
			return err
		}
	}

	return s.roomRepository.Resolve(roomID, lastCommentID)
}

func (s *roomService) SDKGetRoomInfo(ID string) (entities.Room, error) {
	room, err := s.roomRepository.SDKGetRoomInfo(ID)
	if err != nil {
		return entities.Room{}, err
	}

	return room, nil
}

func (s *roomService) UpdateBotState(roomID string, states []int, roomInfo entities.Room) error {
	var roomOptions map[string]interface{}
	json.Unmarshal([]byte(roomInfo.Results.Rooms[0].Options), &roomOptions)

	roomOptions["bot_layer"] = states

	roomOptionsJson, err := json.Marshal(roomOptions)
	if err != nil {
		return err
	}

	err = s.roomRepository.UpdateRoom(roomID, string(roomOptionsJson))
	if err != nil {
		return err
	}

	return nil
}

func (s *roomService) StateExist(room entities.Room) bool {
	return s.roomRepository.StateExist(room)
}

func (s *roomService) QismoRoomInfo(ID string) (viewmodel.QismoRoomInfo, error) {
	room, err := s.roomRepository.QismoRoomInfo(ID)
	if err != nil {
		return viewmodel.QismoRoomInfo{}, err
	}

	return room, nil
}

func (s *roomService) AutoResolveTag(ID string) error {
	return s.roomRepository.TagRoom(ID, os.Getenv("AUTO_RESOLVE_TAG"))
}

func (s *roomService) AutoHandover(ID string) error {
	err := s.roomRepository.ResetBotLayers(ID)
	if err != nil {
		return err
	}
	return s.roomRepository.AutoAssign(ID)
}

func (s *roomService) Handover(ID string) error {
	agents, err := s.multichannelRepository.GetAllAgents(18)
	if err != nil {
		return err
	}

	agentData := agent.GetAvailableRandomlyAgent(agents.Data.Agents)
	err = s.roomRepository.AssignAgent(ID, strconv.Itoa(agentData.ID))
	if err != nil {
		return err
	}

	return nil
}
