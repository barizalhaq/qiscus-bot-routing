package entities

type Multichannel struct {
	appID      string
	adminEmail string
	secret     string
	token      string
}

func NewMultichannel(appID string, adminEmail string, secret string, token string) *Multichannel {
	return &Multichannel{appID, adminEmail, secret, token}
}

func (m *Multichannel) GetSecret() string {
	return m.secret
}

func (m *Multichannel) GetAppID() string {
	return m.appID
}

func (m *Multichannel) GetAdminEmail() string {
	return m.adminEmail
}

func (m *Multichannel) GetToken() string {
	return m.token
}
