package viewmodel

type WebhookRequest struct {
	Payload struct {
		From struct {
			AvatarURL string `json:"avatar_url"`
			Email     string `json:"email"`
			ID        int    `json:"id"`
			IDStr     string `json:"id_str"`
			Name      string `json:"name"`
		} `json:"from"`

		Message struct {
			CommentBeforeID    int         `json:"comment_before_id"`
			CommentBeforeIDStr string      `json:"comment_before_id_str"`
			CreatedAt          string      `json:"created_at"`
			DisableLinkPreview bool        `json:"disable_link_preview"`
			ID                 int         `json:"id"`
			IDStr              string      `json:"id_str"`
			Payload            interface{} `json:"payload"`
			Text               string      `json:"text"`
			Timestamp          string      `json:"timestamp"`
			Type               string      `json:"type"`
			UniqueTempID       string      `json:"unique_temp_id"`
			UnixNanoTimestamp  string      `json:"unix_nano_timestamp"`
			UnixTimestamp      string      `json:"unix_timestamp"`
		} `json:"message"`

		Room struct {
			ID              string      `json:"id"`
			IDStr           string      `json:"id_str"`
			IsPublicChannel bool        `json:"is_public_channel"`
			Name            string      `json:"name"`
			Options         string      `json:"options"`
			Participants    interface{} `json:"participants"`
			RoomAvatar      string      `json:"room_avatar"`
			TopicID         string      `json:"topic_id"`
			TopicIDStr      string      `json:"topic_id_str"`
			Type            string      `json:"type"`
		} `json:"room"`

		Type string `json:"type"`
	} `json:"payload"`
}

type Option struct {
	Channel        string `json:"channel"`
	ChannelDetails struct {
		ChannelID int `json:"channel_id"`
	} `json:"channel_details"`
	IsResolved bool   `json:"is_resolved"`
	IsWaiting  bool   `json:"is_waiting"`
	Source     string `json:"source"`
}

type BotRequestBody struct {
	AdminEmail string `json:"sender_email"`
	Message    string `json:"message"`
	Type       string `json:"type"`
	RoomID     string `json:"room_id"`
}

// type BotRoutingFile struct {
// 	File ``
// }
