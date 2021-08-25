package services

import (
	"bot-routing-engine/entities"
	"bot-routing-engine/entities/viewmodel"
	"bot-routing-engine/repositories"
	"bot-routing-engine/utils/agent"
	"encoding/json"
	"log"
	"os"
	"strconv"
)

type RoomService interface {
	SendBotMessage(roomID string, message string) error
	Resolve(roomID string, lastCommentID string) error
	SDKGetRoomInfo(ID string) (entities.Room, error)
	UpdateBotState(roomID string, states []int, roomInfo entities.Room) (entities.ReturnedUpdatedRoom, error)
	StateExist(room entities.Room) bool
	QismoRoomInfo(ID string) (viewmodel.QismoRoomInfo, error)
	AutoResolveTag(ID string) error
	Handover(ID string) error
	HandoverWithDivision(ID string, division string) error
	Deactivate(ID string) error
	UpdateFormsState(roomID string, states int, roomInfo entities.Room) error
	SaveNewFormData(roomID string, newData map[string]string) error
	UpdateSavedFormData(roomID string, key string, newValue string) error
	SetFormConfirmationStatus(roomID string, status bool, roomInfo entities.Room) error
	SetFormOnEditIndex(roomID string, index int, roomInfo entities.Room) error
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

func (s *roomService) UpdateBotState(roomID string, states []int, roomInfo entities.Room) (entities.ReturnedUpdatedRoom, error) {
	var roomOptions map[string]interface{}
	json.Unmarshal([]byte(roomInfo.Results.Rooms[0].Options), &roomOptions)

	roomOptions["bot_layer"] = states

	roomOptionsJson, err := json.Marshal(roomOptions)
	if err != nil {
		return entities.ReturnedUpdatedRoom{}, err
	}

	updatedRoom, err := s.roomRepository.UpdateRoom(roomID, string(roomOptionsJson))
	if err != nil {
		return entities.ReturnedUpdatedRoom{}, err
	}

	return updatedRoom, nil
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

func (s *roomService) Handover(ID string) error {
	agents, err := s.multichannelRepository.GetAllAgents(100)
	if err != nil {
		return err
	}

	err = s.roomRepository.ResetBotLayers(ID)
	if err != nil {
		return err
	}

	poolAgents, err := s.getPoolAgents()
	if err != nil {
		return err
	}

	anyOnline, agentData := agent.GetAvailableRandomlyAgent(agents.Data.Agents)
	if anyOnline {
		err = s.roomRepository.AssignAgent(ID, strconv.Itoa(agentData.ID))
		if err != nil {
			return err
		}
	} else {
		err = s.assignPoolAgent(ID, poolAgents)
		if err != nil {
			return err
		}
	}

	err = s.Deactivate(ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *roomService) Deactivate(ID string) error {
	return s.roomRepository.ToggleBotInRoom(ID, false)
}

func (s *roomService) assignPoolAgent(roomID string, agents []viewmodel.Agent) error {
	agentData := agent.GetRandomAgent(agents)

	err := s.roomRepository.AssignAgent(roomID, strconv.Itoa(agentData.ID))
	if err != nil {
		return err
	}

	return nil
}

func (s *roomService) HandoverWithDivision(ID string, divisionName string) error {
	err := s.roomRepository.ResetBotLayers(ID)
	if err != nil {
		return err
	}

	divisions, err := s.multichannelRepository.GetAllDivisions()
	if err != nil {
		return err
	}

	divisionData := agent.GetDivisionByName(divisionName, divisions.Data)

	agents, err := s.multichannelRepository.GetAgentsByDivision(strconv.Itoa(divisionData.ID))
	if err != nil {
		return err
	}

	poolAgents, err := s.getPoolAgents()
	if err != nil {
		return err
	}

	anyOnline, agentData := agent.GetAvailableRandomlyAgent(agents.Data)
	if anyOnline {
		err = s.roomRepository.AssignAgent(ID, strconv.Itoa(agentData.ID))
		if err != nil {
			return err
		}
	} else {
		err = s.assignPoolAgent(ID, poolAgents)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *roomService) getPoolAgents() ([]viewmodel.Agent, error) {
	poolAgentDivisionName := os.Getenv("POOL_AGENT_DIVISION")

	divisions, err := s.multichannelRepository.GetAllDivisions()
	if err != nil {
		return []viewmodel.Agent{}, err
	}

	divisionData := agent.GetDivisionByName(poolAgentDivisionName, divisions.Data)
	agents, err := s.multichannelRepository.GetAgentsByDivision(strconv.Itoa(divisionData.ID))
	if err != nil {
		return []viewmodel.Agent{}, err
	}

	return agents.Data, nil
}

func (s *roomService) UpdateFormsState(roomID string, states int, roomInfo entities.Room) error {
	var roomOptions map[string]interface{}
	json.Unmarshal([]byte(roomInfo.Results.Rooms[0].Options), &roomOptions)

	roomOptions["forms_layer_index"] = states

	roomOptionsJson, err := json.Marshal(roomOptions)
	if err != nil {
		return err
	}

	_, err = s.roomRepository.UpdateRoom(roomID, string(roomOptionsJson))
	if err != nil {
		return err
	}

	return nil
}

func (s *roomService) SaveNewFormData(roomID string, newData map[string]string) error {
	userInfo, err := s.roomRepository.GetRoomUserInfo(roomID)
	if err != nil {
		log.Fatalf("Something went wrong: %s", err.Error())
		return err
	}

	if len(userInfo.Data.Extras.UserProperties) > 0 {
		for _, info := range userInfo.Data.Extras.UserProperties {
			if info["key"] == newData["key"] {
				return s.UpdateSavedFormData(roomID, newData["key"], newData["value"])
			}
		}
	}

	apendedUserInfo := append(userInfo.Data.Extras.UserProperties, newData)

	data := map[string][]map[string]string{
		"user_properties": apendedUserInfo,
	}
	err = s.roomRepository.SaveUserInfo(roomID, data)
	if err != nil {
		log.Fatalf("Something went wrong: %s", err.Error())
		return err
	}

	return nil
}

func (s *roomService) UpdateSavedFormData(roomID string, key string, newValue string) error {
	userInfo, err := s.roomRepository.GetRoomUserInfo(roomID)
	if err != nil {
		log.Fatalf("Something went wrong: %s", err.Error())
		return err
	}

	existingInformation := userInfo.Data.Extras.UserProperties
	for _, info := range existingInformation {
		if info["key"] == key {
			info["value"] = newValue
		}
	}

	data := map[string][]map[string]string{
		"user_properties": existingInformation,
	}

	err = s.roomRepository.SaveUserInfo(roomID, data)
	if err != nil {
		log.Fatalf("Something went wrong: %s", err.Error())
		return err
	}

	return nil
}

func (s *roomService) SetFormConfirmationStatus(roomID string, status bool, roomInfo entities.Room) error {
	var roomOptions map[string]interface{}
	json.Unmarshal([]byte(roomInfo.Results.Rooms[0].Options), &roomOptions)

	roomOptions["form_confirmed"] = status

	roomOptionsJson, err := json.Marshal(roomOptions)
	if err != nil {
		return err
	}

	_, err = s.roomRepository.UpdateRoom(roomID, string(roomOptionsJson))
	if err != nil {
		return err
	}

	return nil
}

func (s *roomService) SetFormOnEditIndex(roomID string, index int, roomInfo entities.Room) error {
	var roomOptions map[string]interface{}
	json.Unmarshal([]byte(roomInfo.Results.Rooms[0].Options), &roomOptions)

	roomOptions["form_on_edit_index"] = index

	roomOptionsJson, err := json.Marshal(roomOptions)
	if err != nil {
		return err
	}

	_, err = s.roomRepository.UpdateRoom(roomID, string(roomOptionsJson))
	if err != nil {
		return err
	}

	return nil
}
