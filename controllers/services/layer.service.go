package services

import (
	"bot-routing-engine/entities/viewmodel"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

type LayerService interface {
	GetLayer(source int) (viewmodel.Layer, error)
	DetermineLayer(state string, states []int, layer viewmodel.Layer) (viewmodel.Layer, error)
}

type layerService struct {
}

func NewLayerService() *layerService {
	return &layerService{}
}

func (ls *layerService) GetLayer(source int) (viewmodel.Layer, error) {
	filePath := fmt.Sprintf("./layer/%d.json", source)

	if os.Getenv("ALL_IN_ONE_JSON_ROUTE") == "true" {
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
		if layer.Handover || layer.Resolve {
			return layer, nil
		}
	}

	if layer.Input {
		return layer.Options[0], nil
	}

	option, err := strconv.Atoi(state)
	if err != nil {
		return layer, errors.New("mohon untuk menjawab pilihan layanan hanya dalam format angka (misal: ketik '1'), sesuai dengan pilihan yang disediakan. Terima kasih")
	}

	if option <= 0 || option > len(layer.Options) {
		return layer, errors.New("mohon untuk menjawab pilihan layanan hanya dalam format angka (misal: ketik '1'), sesuai dengan pilihan yang disediakan. Terima kasih")
	}

	layer = layer.Options[option-1]

	return layer, nil
}
