package msclient

import (
	"context"
	"fmt"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go-core/authentication"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

const (
	AuthHost     = "https://login.microsoftonline.com/"
	GraphAPIHost = "https://graph.microsoft.com"
)

const (
	DefaultMicrosoftGraphScope = "https://graph.microsoft.com/.default"
)

func NewMicrosoftGraph(conf Config) (*MicrosoftGraph, error) {
	oauthConf := &oauth2.Config{
		ClientID:     conf.ClientId,
		ClientSecret: conf.ClientSecret,
		RedirectURL:  conf.RedirectURL,
		Endpoint:     microsoft.AzureADEndpoint(conf.TenantId),
		Scopes: []string{
			DefaultMicrosoftGraphScope,
		},
	}

	msCred, err := confidential.NewCredFromSecret(conf.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("could not create microsoft cred: %v", err)
	}

	app, err := confidential.New(fmt.Sprintf("%s%s", AuthHost, conf.TenantId), conf.ClientId, msCred)
	if err != nil {
		return nil, fmt.Errorf("could not create microsoft confidential client: %v", err)
	}
	return &MicrosoftGraph{
		conf:      conf,
		oauthConf: oauthConf,
		app:       app,
	}, nil
}

type MicrosoftGraph struct {
	conf      Config
	oauthConf *oauth2.Config
	app       confidential.Client
}

func microsoftGraphClient(ctx context.Context, token Token) (*msgraphsdk.GraphServiceClient, error) {
	oauthToken, err := token.refresh(ctx)
	if err != nil {
		return nil, err
	}
	auth, err := authentication.NewAzureIdentityAuthenticationProvider(azureTokenCredential{token: oauthToken})
	if err != nil {
		return nil, err
	}
	adapter, err := msgraphsdk.NewGraphRequestAdapter(auth)
	if err != nil {
		return nil, err
	}
	return msgraphsdk.NewGraphServiceClient(adapter), nil
}

func (c *MicrosoftGraph) AuthUrl(ctx context.Context, scopes ...string) (string, error) {
	u, err := c.app.AuthCodeURL(ctx, c.conf.ClientId, c.conf.RedirectURL, scopes)
	if err != nil {
		return "", fmt.Errorf("could not create auth code url: %v", err)
	}
	return u, nil
}

func (c *MicrosoftGraph) AuthUrlCallback(ctx context.Context, code string) (Token, error) {
	// 使用授权码交换访问令牌
	tk, err := c.oauthConf.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange token failed:%v", err)
	}
	return &token{
		oauth:       tk,
		oauthConfig: c.oauthConf,
	}, nil

}
