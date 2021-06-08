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
