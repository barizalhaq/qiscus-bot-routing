package agent

import (
	"bot-routing-engine/entities/viewmodel"
	"math/rand"
)

func GetAvailableRandomlyAgent(agents []viewmodel.Agent) viewmodel.Agent {
	var sortedAgents []viewmodel.Agent
	for _, agent := range agents {
		if agent.IsAvailable {
			sortedAgents = append(sortedAgents, agent)
		}
	}

	var randomIndex int
	if len(sortedAgents) > 0 {
		randomIndex = rand.Intn(len(sortedAgents)-0) + 0

		return sortedAgents[randomIndex]
	}

	randomIndex = rand.Intn(len(agents)-0) + 0
	return agents[randomIndex]
}
