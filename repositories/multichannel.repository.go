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
	"os"
)

type MultichannelRepository interface {
	SendBotMessage(roomID string, message string) error
	GetAllAgents(limit int) (viewmodel.AgentsResponse, error)
	OfficeHour() (viewmodel.OfficeHourResp, error)
	GetAllDivisions() (viewmodel.Divisions, error)
	GetAgentsByDivision(divisionID string) (viewmodel.AgentsByDivision, error)
}

type multichannelRepository struct {
	qismoURL      string
	multichannel  *entities.Multichannel
	outbondLogger *log.Logger
}

func NewMultichannelRepository(multichannel *entities.Multichannel, outbondLogger *log.Logger) *multichannelRepository {
	return &multichannelRepository{
		qismoURL:      os.Getenv("QISMO_BASE_URL"),
		multichannel:  multichannel,
		outbondLogger: outbondLogger,
	}
}

func (r *multichannelRepository) SendBotMessage(roomID string, message string) error {
	url := fmt.Sprintf("%s/%s/bot", r.qismoURL, r.multichannel.GetAppID())
	method := "POST"
	payload := viewmodel.BotRequestBody{
		AdminEmail: r.multichannel.GetAdminEmail(),
		Message:    message,
		Type:       "text",
		RoomID:     roomID,
	}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("QISCUS_SDK_SECRET", r.multichannel.GetSecret())

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	logger.WriteOutbondLog(r.outbondLogger, res, string(body), "")

	return nil
}

func (r *multichannelRepository) GetAllAgents(limit int) (viewmodel.AgentsResponse, error) {
	url := fmt.Sprintf("%s/api/v2/admin/agents?limit=%d", r.qismoURL, limit)
	method := "GET"

	client := &http.Client{}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return viewmodel.AgentsResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", r.multichannel.GetToken())
	req.Header.Set("Qiscus-App-Id", r.multichannel.GetAppID())

	resp, err := client.Do(req)
	if err != nil {
		return viewmodel.AgentsResponse{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return viewmodel.AgentsResponse{}, err
	}

	logger.WriteOutbondLog(r.outbondLogger, resp, string(body), "")

	var agentsResp viewmodel.AgentsResponse
	json.Unmarshal(body, &agentsResp)

	return agentsResp, nil
}

func (r *multichannelRepository) OfficeHour() (viewmodel.OfficeHourResp, error) {
	url := fmt.Sprintf("%s/api/v1/admin/office_hours", r.qismoURL)
	method := "GET"

	client := &http.Client{}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return viewmodel.OfficeHourResp{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", r.multichannel.GetToken())
	req.Header.Set("Qiscus-App-Id", r.multichannel.GetAppID())

	resp, err := client.Do(req)
	if err != nil {
		return viewmodel.OfficeHourResp{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return viewmodel.OfficeHourResp{}, err
	}

	logger.WriteOutbondLog(r.outbondLogger, resp, string(body), "")

	var officeHour viewmodel.OfficeHourResp
	json.Unmarshal(body, &officeHour)

	return officeHour, nil
}

func (r *multichannelRepository) GetAllDivisions() (viewmodel.Divisions, error) {
	url := fmt.Sprintf("%s/api/v2/divisions", r.qismoURL)
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return viewmodel.Divisions{}, err
	}

	qParams := req.URL.Query()
	qParams.Add("limit", "100")
	req.URL.RawQuery = qParams.Encode()

	req.Header.Set("Authorization", r.multichannel.GetToken())

	resp, err := client.Do(req)
	if err != nil {
		return viewmodel.Divisions{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return viewmodel.Divisions{}, err
	}

	logger.WriteOutbondLog(r.outbondLogger, resp, string(body), "")

	var divisions viewmodel.Divisions
	json.Unmarshal(body, &divisions)

	return divisions, nil
}

func (r *multichannelRepository) GetAgentsByDivision(divisionID string) (viewmodel.AgentsByDivision, error) {
	url := fmt.Sprintf("%s/api/v2/admin/agents/by_division", r.qismoURL)

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return viewmodel.AgentsByDivision{}, err
	}

	qParams := req.URL.Query()
	qParams.Add("limit", "100")
	qParams.Add("division_ids[]", divisionID)
	req.URL.RawQuery = qParams.Encode()

	req.Header.Set("Authorization", r.multichannel.GetToken())

	resp, err := client.Do(req)
	if err != nil {
		return viewmodel.AgentsByDivision{}, nil
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return viewmodel.AgentsByDivision{}, nil
	}

	logger.WriteOutbondLog(r.outbondLogger, resp, string(body), "")

	var agents viewmodel.AgentsByDivision
	json.Unmarshal(body, &agents)

	return agents, nil
}
