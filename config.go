package msclient

type Config struct {
	TenantId     string `json:"tenant_id"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
}
