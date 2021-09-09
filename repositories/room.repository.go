package repositories

import (
	"bot-routing-engine/entities"
	"bot-routing-engine/entities/viewmodel"
	"bot-routing-engine/utils/logger"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type RoomRepository interface {
	SDKGetRoomInfo(ID string) (entities.Room, error)
	StateExist(room entities.Room) bool
	UpdateRoom(ID string, options string) (entities.ReturnedUpdatedRoom, error)
	Resolve(ID string, lastCommentID string) error
	QismoRoomInfo(ID string) (viewmodel.QismoRoomInfo, error)
	ResetBotLayers(ID string) error
	TagRoom(ID string, tag string) error
	AssignAgent(ID string, agentID string) error
	ToggleBotInRoom(ID string, activate bool) error
	FormStateExist(room entities.Room) bool
	GetRoomUserInfo(ID string) (entities.UserInfo, error)
	SaveUserInfo(ID string, data map[string][]map[string]string) error
	Confirming(room entities.Room) bool
	DeleteRoomOption(ID string, key string) error
}

type roomRepository struct {
	sdkURL        string
	qismoUrl      string
	multichannel  *entities.Multichannel
	outbondLogger *log.Logger
}

func NewRoomRepository(multichannel *entities.Multichannel, outbondLogger *log.Logger) *roomRepository {
	return &roomRepository{
		sdkURL:        os.Getenv("SDK_URL"),
		qismoUrl:      os.Getenv("QISMO_BASE_URL"),
		multichannel:  multichannel,
		outbondLogger: outbondLogger,
	}
}

func (r *roomRepository) SDKGetRoomInfo(ID string) (entities.Room, error) {
	var room entities.Room

	url := r.sdkURL + "/rest/get_rooms_info?room_ids%5B%5D=" + ID
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return room, err
	}

	req.Header.Set("QISCUS-SDK-SECRET", r.multichannel.GetSecret())
	req.Header.Set("QISCUS-SDK-APP-ID", r.multichannel.GetAppID())

	res, err := client.Do(req)
	if err != nil {
		return room, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return room, err
	}

	json.Unmarshal(body, &room)

	logger.WriteOutbondLog(r.outbondLogger, res, string(body), "")

	return room, nil
}

func (r *roomRepository) StateExist(room entities.Room) bool {
	var roomOptions map[string]string
	json.Unmarshal([]byte(room.Results.Rooms[0].Options), &roomOptions)

	_, ok := roomOptions["bot_layer"]

	return ok
}

func (r *roomRepository) UpdateRoom(ID string, options string) (entities.ReturnedUpdatedRoom, error) {
	url := fmt.Sprintf("%s/rest/update_room", r.sdkURL)
	method := "POST"
	payload, err := json.Marshal(map[string]string{
		"room_id":      ID,
		"room_options": options,
	})
	if err != nil {
		return entities.ReturnedUpdatedRoom{}, err
	}

	client := &http.Client{}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return entities.ReturnedUpdatedRoom{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("QISCUS-SDK-APP-ID", r.multichannel.GetAppID())
	req.Header.Set("QISCUS-SDK-SECRET", r.multichannel.GetSecret())

	resp, err := client.Do(req)
	if err != nil {
		return entities.ReturnedUpdatedRoom{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return entities.ReturnedUpdatedRoom{}, err
	}

	var updatedRoom entities.ReturnedUpdatedRoom
	json.Unmarshal(body, &updatedRoom)

	logger.WriteOutbondLog(r.outbondLogger, resp, string(body), string(payload))

	return updatedRoom, nil
}

func (r *roomRepository) Resolve(ID string, lastCommentID string) error {
	url := fmt.Sprintf("%s/api/v1/admin/service/mark_as_resolved", r.qismoUrl)
	method := "POST"
	payload, err := json.Marshal(map[string]string{
		"room_id":         ID,
		"last_comment_id": lastCommentID,
	})
	if err != nil {
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Qiscus-App-Id", r.multichannel.GetAppID())
	req.Header.Set("Qiscus-Secret-Key", r.multichannel.GetSecret())

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	logger.WriteOutbondLog(r.outbondLogger, resp, string(body), string(payload))

	return nil
}

func (r *roomRepository) QismoRoomInfo(ID string) (viewmodel.QismoRoomInfo, error) {
	url := fmt.Sprintf("%s/api/v2/customer_rooms/%s", r.qismoUrl, ID)
	method := "GET"

	client := &http.Client{}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return viewmodel.QismoRoomInfo{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", r.multichannel.GetToken())
	req.Header.Set("Qiscus-App-Id", r.multichannel.GetAppID())

	resp, err := client.Do(req)
	if err != nil {
		return viewmodel.QismoRoomInfo{}, err
	}

	defer resp.Body.Close()

	var room viewmodel.QismoRoomInfo
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return viewmodel.QismoRoomInfo{}, err
	}
	json.Unmarshal(body, &room)

	logger.WriteOutbondLog(r.outbondLogger, resp, string(body), "")

	return room, nil
}

func (r *roomRepository) ResetBotLayers(ID string) error {
	r.DeleteRoomOption(ID, "bot_layer")
	r.DeleteRoomOption(ID, "forms_layer_index")
	r.DeleteRoomOption(ID, "nested_form_keys")
	r.DeleteRoomOption(ID, "form_confirming")

	return nil
}

func (r *roomRepository) TagRoom(ID string, tag string) error {
	apiUrl := fmt.Sprintf("%s/api/v1/room_tag/create", r.qismoUrl)
	method := "POST"

	formData := url.Values{}
	formData.Set("room_id", ID)
	formData.Set("tag", os.Getenv("AUTO_RESOLVE_TAG"))

	client := &http.Client{}

	req, err := http.NewRequest(method, apiUrl, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	req.Header.Set("Authorization", r.multichannel.GetToken())
	req.Header.Set("Qiscus-App-Id", r.multichannel.GetAppID())

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	logger.WriteOutbondLog(r.outbondLogger, resp, string(body), formData.Encode())

	return nil
}

func (r *roomRepository) AssignAgent(ID string, agentID string) error {
	apiUrl := fmt.Sprintf("%s/api/v1/admin/service/assign_agent", r.qismoUrl)
	method := "POST"

	formData := url.Values{}
	formData.Set("agent_id", agentID)
	formData.Set("room_id", ID)

	client := &http.Client{}
	req, err := http.NewRequest(method, apiUrl, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Qiscus-App-Id", r.multichannel.GetAppID())
	req.Header.Set("Qiscus-Secret-Key", r.multichannel.GetSecret())

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	logger.WriteOutbondLog(r.outbondLogger, resp, string(body), formData.Encode())

	return nil
}

func (r *roomRepository) ToggleBotInRoom(ID string, activate bool) error {
	apiUrl := fmt.Sprintf("%s/bot/%s/activate", r.qismoUrl, ID)
	method := "POST"

	formData := url.Values{}
	formData.Set("is_active", strconv.FormatBool(activate))

	client := &http.Client{}

	req, err := http.NewRequest(method, apiUrl, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", r.multichannel.GetToken())

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	logger.WriteOutbondLog(r.outbondLogger, resp, string(body), formData.Encode())

	return nil
}

func (r *roomRepository) FormStateExist(room entities.Room) bool {
	var roomOptions map[string]string
	json.Unmarshal([]byte(room.Results.Rooms[0].Options), &roomOptions)

	_, ok := roomOptions["forms_layer_index"]

	return ok
}

func (r *roomRepository) GetRoomUserInfo(ID string) (entities.UserInfo, error) {
	url := fmt.Sprintf("%s/api/v1/qiscus/room/%s/user_info", r.qismoUrl, ID)
	method := "GET"

	client := &http.Client{}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Fatalf("Something went wrong: %s", err.Error())
		return entities.UserInfo{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", r.multichannel.GetToken())

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Something went wrong: %s", err.Error())
		return entities.UserInfo{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Something went wrong: %s", err.Error())
		return entities.UserInfo{}, err
	}

	var roomUserInfo entities.UserInfo
	json.Unmarshal(body, &roomUserInfo)

	return roomUserInfo, nil
}

func (r *roomRepository) SaveUserInfo(ID string, data map[string][]map[string]string) error {
	apiUrl := fmt.Sprintf("%s/api/v1/qiscus/room/%s/user_info", r.qismoUrl, ID)
	method := "POST"
	payload, _ := json.Marshal(data)

	client := &http.Client{}

	req, err := http.NewRequest(method, apiUrl, bytes.NewBuffer(payload))
	if err != nil {
		log.Fatalf("Something went wrong: %s", err.Error())
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Qiscus-App-Id", r.multichannel.GetAppID())
	req.Header.Set("Qiscus-Secret-Key", r.multichannel.GetSecret())

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Something went wrong: %s", err.Error())
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Something went wrong: %s", err.Error())
		return err
	}

	logger.WriteOutbondLog(r.outbondLogger, resp, string(body), string(payload))

	return nil
}

func (r *roomRepository) Confirming(room entities.Room) bool {
	var roomOptions map[string]bool
	json.Unmarshal([]byte(room.Results.Rooms[0].Options), &roomOptions)

	confirming, ok := roomOptions["form_confirming"]

	if ok {
		return confirming
	}

	return ok
}

func (r *roomRepository) DeleteRoomOption(ID string, key string) error {
	roomInfo, err := r.SDKGetRoomInfo(ID)
	if err != nil {
		return err
	}

	var roomOptions map[string]interface{}
	json.Unmarshal([]byte(roomInfo.Results.Rooms[0].Options), &roomOptions)

	delete(roomOptions, key)
	options, err := json.Marshal(roomOptions)
	if err != nil {
		return err
	}

	_, err = r.UpdateRoom(ID, string(options))
	if err != nil {
		return err
	}

	return nil
}
