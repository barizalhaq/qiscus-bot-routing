package agent

import (
	"bot-routing-engine/entities/viewmodel"
	"math/rand"
	"time"
)

func GetAvailableRandomlyAgent(agents []viewmodel.Agent) (bool, viewmodel.Agent) {
	var onlineAgents []viewmodel.Agent
	for _, agent := range agents {
		if agent.IsAvailable {
			onlineAgents = append(onlineAgents, agent)
		}
	}

	if len(onlineAgents) > 0 {
		rand.Seed(time.Now().Unix())
		randomIndex := rand.Intn(len(onlineAgents))
		return true, onlineAgents[randomIndex]
	}

	return false, viewmodel.Agent{}
}

func GetDivisionByName(divisionName string, divisions []viewmodel.Division) viewmodel.Division {
	for _, division := range divisions {
		if divisionName == division.Name {
			return division
		}
	}

	return viewmodel.Division{}
}

func GetRandomAgent(agents []viewmodel.Agent) viewmodel.Agent {
	rand.Seed(time.Now().Unix())
	randomIndex := rand.Intn(len(agents))

	return agents[randomIndex]
}
