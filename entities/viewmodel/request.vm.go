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

type Draft struct {
	Room    *WebhookRequest
	Layer   Layer
	Message string
}

type QismoRoomInfo struct {
	Data struct {
		ChannelID int    `json:"channel_id"`
		ID        int    `json:"id"`
		Name      string `json:"name"`
	} `json:"data"`
	Status int `json:"status"`
}

type AgentsResponse struct {
	Data struct {
		Agents []Agent `json:"agents"`
	} `json:"data"`
	Meta   Meta `json:"meta"`
	Status int  `json:"status"`
}

type Agent struct {
	AvatarUrl     string    `json:"avatar_url"`
	CustomerCount int       `json:"current_customer_count"`
	Email         string    `json:"email"`
	ForceOffline  bool      `json:"force_offline"`
	ID            int       `json:"id"`
	IsAvailable   bool      `json:"is_available"`
	LastLogin     string    `json:"last_login"`
	Name          string    `json:"name"`
	SdkEmail      string    `json:"sdk_email"`
	SdkKey        string    `json:"sdk_key"`
	Type          int       `json:"type"`
	TypeStr       string    `json:"type_as_string"`
	Channels      []Channel `json:"user_channels"`
	Roles         []Role    `json:"user_roles"`
}

type Channel struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Meta struct {
	PerPage    int `json:"per_page"`
	TotalCount int `json:"total_count"`
}

type OfficeHourResp struct {
	Data struct {
		OnlineMessage          string       `json:"online_message"`
		OfflineMessage         string       `json:"offline_message"`
		Timezone               string       `json:"timezone"`
		SendOnlineIfResolved   bool         `json:"send_online_if_resolved"`
		SendOfflineEachMessage bool         `json:"send_offline_each_message"`
		OfficeHours            []OfficeHour `json:"office_hours"`
	} `json:"data"`
}

type OfficeHour struct {
	ID        int    `json:"id"`
	AppID     int    `json:"app_id"`
	Day       int    `json:"day"`
	Starttime string `json:"starttime"`
	Endtime   string `json:"endtime"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type WebhookResolved struct {
	Customer struct {
		AdditionalInfo []interface{} `json:"additional_info"`
		Avatar         string        `json:"avatar"`
		Name           string        `json:"name"`
		UserID         string        `json:"user_id"`
	} `json:"customer"`
	ResolvedBy struct {
		Email       string `json:"email"`
		ID          int    `json:"id"`
		IsAvailable bool   `json:"is_available"`
		Name        string `json:"name"`
		Type        string `json:"type"`
	} `json:"resolved_by"`
	Service struct {
		FirstCommentID string `json:"first_comment_id"`
		ID             int    `json:"id"`
		IsResolved     bool   `json:"is_resolved"`
		LastCommentID  int    `json:"last_comment_id"`
		Notes          string `json:"notes"`
		RoomID         string `json:"room_id"`
		Source         string `json:"source"`
	} `json:"service"`
}
