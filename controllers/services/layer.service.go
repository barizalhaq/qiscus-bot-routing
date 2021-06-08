package services

import (
	"bot-routing-engine/entities/viewmodel"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type LayerService interface {
	GetLayer(source int) (viewmodel.Layer, error)
}

type layerService struct{
}

func NewLayerService() *layerService {
	return &layerService{}
}

func (ls *layerService) GetLayer(source int) (viewmodel.Layer, error) {
	jsonFile, err := os.Open(fmt.Sprintf("./layer/%d.json", source))
	if err != nil {
		return viewmodel.Layer{}, err
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)


	var layer viewmodel.Layer
	json.Unmarshal(byteValue, &layer)

	return layer, nil
}