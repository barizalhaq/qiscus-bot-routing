package entities

type Room struct {
	Results struct {
		Rooms []Rooms `json:"rooms"`
	} `json:"results"`
	Status int `json:"status"`
}

type Rooms struct {
	AvatarURL string `json:"room_avatar_url"`
	ChannelID string `json:"room_channel_id"`
	ID        string `json:"room_id"`
	Name      string `json:"room_name"`
	Options   string `json:"room_options"`
	Type      string `json:"type"`
}

type ReturnedUpdatedRoom struct {
	Results struct {
		Changed bool  `json:"changed"`
		Room    Rooms `json:"room"`
	} `json:"results"`
	Status int `json:"status"`
}

type UserInfo struct {
	Data struct {
		Extras struct {
			UserProperties []map[string]string `json:"user_properties"`
		} `json:"extras"`
		FirstInitiated         string      `json:"first_initiated"`
		FirstAgentResponseTime interface{} `json:"first_agent_response_time"`
		UserID                 string      `json:"user_id"`
		ChannelID              int         `json:"channel_id"`
		IsBlocked              bool        `json:"is_blocked"`
		ChannelName            string      `json:"channel_name"`
		Channel                interface{} `json:"channel"`
	} `json:"data"`
}
