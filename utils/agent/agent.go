package agent

import (
	"bot-routing-engine/entities/viewmodel"
	"math/rand"
	"time"
)

func GetAvailableRandomlyAgent(agents []viewmodel.Agent, channelID int) (bool, viewmodel.Agent) {
	var channelAgents []viewmodel.Agent
	for _, agent := range agents {
		for _, agentChannel := range agent.Channels {
			if agentChannel.ID == channelID {
				channelAgents = append(channelAgents, agent)
			}
		}
	}

	var onlineAgents []viewmodel.Agent
	for _, agent := range channelAgents {
		if agent.IsAvailable && agent.TypeStr == "agent" {
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

func GetRandomAgent(agents []viewmodel.Agent, channelID int) viewmodel.Agent {
	var channelAgents []viewmodel.Agent
	for _, agent := range agents {
		for _, agentChannel := range agent.Channels {
			if agentChannel.ID == channelID {
				channelAgents = append(channelAgents, agent)
			}
		}
	}

	var agentsOnly []viewmodel.Agent
	for _, agent := range channelAgents {
		if agent.TypeStr == "agent" {
			agentsOnly = append(agentsOnly, agent)
		}
	}

	rand.Seed(time.Now().Unix())
	randomIndex := rand.Intn(len(agentsOnly))

	return agentsOnly[randomIndex]
}
