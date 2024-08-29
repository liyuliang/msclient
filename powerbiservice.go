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
	PowerBIAPIHost = "https://api.powerbi.com"
)

func NewPowerBIService(conf Config) *PowerBIService {
	oauthConf := &oauth2.Config{
		ClientID:     conf.ClientId,
		ClientSecret: conf.ClientSecret,
		RedirectURL:  conf.RedirectURL,
		Endpoint:     microsoft.AzureADEndpoint(conf.TenantId),
		Scopes: []string{
			DefaultMicrosoftGraphScope,
		},
	}
	return &PowerBIService{
		conf:      conf,
		oauthConf: oauthConf,
	}
}

type PowerBIService struct {
	conf      Config
	oauthConf *oauth2.Config
}

/*
RefreshDatasetInGroup Refresh Dataset In Group
https://learn.microsoft.com/en-us/rest/api/power-bi/datasets/refresh-dataset-in-group#code-try-0
*/
func (c *PowerBIService) RefreshDatasetInGroup(ctx context.Context, token Token, groupId string, datasetId string) ([]byte, error) {
	oauth, err := token.refresh(ctx)
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf(`%s/v1.0/myorg/groups/%s/datasets/%s/refreshes`, PowerBIAPIHost, groupId, datasetId)

	req, err := http.NewRequest(http.MethodPost, u, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	oauth.SetAuthHeader(req)

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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("error status: %d , %s", resp.StatusCode, body)
	}
	return body, nil
}

func (c *PowerBIService) ExchangeToken() (Token, error) {

	var (
		resource = "https://analysis.windows.net/powerbi/api" // Resource URI for Power BI API
		scope    = "https://analysis.windows.net/powerbi/api/.default"
	)
	//scope = `App.Read.All Capacity.Read.All Capacity.ReadWrite.All Content.Create Dashboard.Read.All Dashboard.ReadWrite.All Dataflow.Read.All Dataflow.ReadWrite.All Dataset.Read.All Dataset.ReadWrite.All Gateway.Read.All Gateway.ReadWrite.All Item.Execute.All Item.ExternalDataShare.All Item.ReadWrite.All Item.Reshare.All OneLake.Read.All OneLake.ReadWrite.All Pipeline.Deploy Pipeline.Read.All Pipeline.ReadWrite.All Report.ReadWrite.All Reprt.Read.All StorageAccount.Read.All StorageAccount.ReadWrite.All Tenant.Read.All Tenant.ReadWrite.All UserState.ReadWrite.All Workspace.GitCommit.All Workspace.GitUpdate.All Workspace.Read.All Workspace.ReadWrite.All`

	u := fmt.Sprintf("%s%s", AuthHost, c.conf.TenantId) + "/oauth2/token"

	payload := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     c.conf.ClientId,
		"client_secret": c.conf.ClientSecret,
		"resource":      resource,
		"scope":         scope,
	}

	// Create HTTP client
	client := resty.New()

	// Make request to Azure AD token endpoint
	resp, err := client.R().
		SetFormData(payload).Post(u)

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
