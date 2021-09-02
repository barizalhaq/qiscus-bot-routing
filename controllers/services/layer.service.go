package services

import (
	"bot-routing-engine/entities/viewmodel"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

type LayerService interface {
	GetLayer(source int) (viewmodel.Layer, error)
	DetermineLayer(state string, states []int, layer viewmodel.Layer) (viewmodel.Layer, error)
	GetFormConfirmationOption(state string, layer viewmodel.Layer) (bool, string, error)
}

type layerService struct {
}

func NewLayerService() *layerService {
	return &layerService{}
}

func (ls *layerService) GetLayer(source int) (viewmodel.Layer, error) {
	layerURL, layerURLExist := os.LookupEnv(fmt.Sprintf("%v_LAYER_URL", source))
	if layerURLExist {
		return ls.getLayerFromURL(source, layerURL)
	}

	filePath := fmt.Sprintf("./layer/%d.json", source)

	if os.Getenv("ALL_IN_ONE_JSON_ROUTE") == "true" {
		layerURL, layerURLExist = os.LookupEnv("LAYER_URL")
		if layerURLExist {
			return ls.getLayerFromURL(source, layerURL)
		}
		filePath = "./layer/layer.json"
	}

	jsonFile, err := os.Open(filePath)
	if err != nil {
		return viewmodel.Layer{}, err
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var layer viewmodel.Layer
	json.Unmarshal(byteValue, &layer)

	return layer, nil
}

func (s *layerService) getLayerFromURL(source int, url string) (viewmodel.Layer, error) {
	resp, err := http.Get(url)
	if err != nil {
		return viewmodel.Layer{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return viewmodel.Layer{}, fmt.Errorf("unexpected http GET status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return viewmodel.Layer{}, err
	}

	var layer viewmodel.Layer
	json.Unmarshal(body, &layer)

	return layer, nil
}

func (s *layerService) getLatestLayer(states []int, layer viewmodel.Layer) viewmodel.Layer {
	for _, state := range states {
		layer = layer.Options[state-1]
	}
	return layer
}

func (s *layerService) DetermineLayer(state string, states []int, layer viewmodel.Layer) (viewmodel.Layer, error) {
	if len(states) > 0 {
		if prevLayerKeypad, prevLayerKeypadEnable := os.LookupEnv("RETURN_PREVIOUS_LAYER_KEYPAD"); prevLayerKeypadEnable &&
			state == prevLayerKeypad {
			removedLastState := states[:len(states)-1]
			return s.getLatestLayer(removedLastState, layer), nil
		}

		layer = s.getLatestLayer(states, layer)
		if layer.Handover || layer.Resolve || layer.AddAdditionalInformation {
			return layer, nil
		}
	}

	if layer.Input {
		return layer.Options[0], nil
	}

	option, err := strconv.Atoi(state)
	if err != nil {
		return layer, errors.New(os.Getenv("FALLBACK_MESSAGE"))
	}

	if option <= 0 || option > len(layer.Options) {
		return layer, errors.New(os.Getenv("FALLBACK_MESSAGE"))
	}

	layer = layer.Options[option-1]

	return layer, nil
}

func (s *layerService) GetFormConfirmationOption(state string, layer viewmodel.Layer) (bool, string, error) {
	confirmationOptions := layer.AdditionalInformation.FormsConfirmation.Options

	option, err := strconv.Atoi(state)
	if err != nil || option <= 0 || option > len(confirmationOptions) {
		return false, "", errors.New(os.Getenv("FALLBACK_MESSAGE"))
	}

	fmt.Println(confirmationOptions[option-1].Confirmed)

	return confirmationOptions[option-1].Confirmed, confirmationOptions[option-1].Message, nil
}
