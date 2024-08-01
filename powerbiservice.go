package msclient

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
	"io/ioutil"
	"net/http"
)

const (
	powerBIAPIHost = "https://api.powerbi.com"
)

type PowerBIService interface {
	ExchangeToken() (Token, error)
	RefreshDatasetInGroup(ctx context.Context, token Token, groupId string, datasetId string) ([]byte, error)
}

func NewPowerBIService(conf Config) PowerBIService {
	oauthConf := &oauth2.Config{
		ClientID:     conf.ClientId,
		ClientSecret: conf.ClientSecret,
		RedirectURL:  conf.RedirectURL,
		Endpoint:     microsoft.AzureADEndpoint(conf.TenantId),
		Scopes: []string{
			defaultMicrosoftGraphScope,
		},
	}
	return &powerBIService{
		conf:      conf,
		oauthConf: oauthConf,
	}
}

type powerBIService struct {
	conf      Config
	oauthConf *oauth2.Config
}

/*
RefreshDatasetInGroup Refresh Dataset In Group
https://learn.microsoft.com/en-us/rest/api/power-bi/datasets/refresh-dataset-in-group#code-try-0
*/
func (c *powerBIService) RefreshDatasetInGroup(ctx context.Context, token Token, groupId string, datasetId string) ([]byte, error) {
	oauth, err := token.refresh(ctx)
	if err != nil {
		return nil, err
	}

	header := http.Header{}
	header.Set("Authorization", fmt.Sprintf("Bearer %s", oauth.AccessToken))

	u := fmt.Sprintf(`%s/v1.0/myorg/groups/%s/datasets/%s/refreshes`, powerBIAPIHost, groupId, datasetId)

	req, err := http.NewRequest(http.MethodPost, u, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header = header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	return body, nil
}

func (c *powerBIService) ExchangeToken() (Token, error) {

	var (
		resource = "https://analysis.windows.net/powerbi/api" // Resource URI for Power BI API
		scope    = "https://analysis.windows.net/powerbi/api/.default"
	)

	u := fmt.Sprintf("%s%s", authHost, c.conf.TenantId) + "/oauth2/token"

	// Create HTTP client
	client := resty.New()

	// Make request to Azure AD token endpoint
	resp, err := client.R().
		SetFormData(map[string]string{
			"grant_type":    "client_credentials",
			"client_id":     c.conf.ClientId,
			"client_secret": c.conf.ClientSecret,
			"resource":      resource,
			"scope":         scope,
		}).Post(u)

	if err != nil {
		return nil, err
	}

	t := &oauth2.Token{}
	if err := json.Unmarshal(resp.Body(), t); err != nil {
		return nil, err
	}

	return &token{
		oauth:       t,
		oauthConfig: c.oauthConf,
	}, nil
}
