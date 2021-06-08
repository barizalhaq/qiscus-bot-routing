package repositories

import (
	"bot-routing-engine/entities"
	"bot-routing-engine/utils/logger"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type RoomRepository interface {
	GetRoomInfo(ID string) (entities.Room, error)
	StateExist(room entities.Room) bool
}

type roomRepository struct {
	sdkURL        string
	multichannel  *entities.Multichannel
	outbondLogger *log.Logger
}

func NewRoomRepository(multichannel *entities.Multichannel, outbondLogger *log.Logger) *roomRepository {
	return &roomRepository{
		sdkURL:        os.Getenv("SDK_URL"),
		multichannel:  multichannel,
		outbondLogger: outbondLogger,
	}
}

func (r *roomRepository) GetRoomInfo(ID string) (entities.Room, error) {
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

	_, ok := roomOptions["bot_state"]

	return ok
}
