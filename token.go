package msclient

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"net/http"
)

type Token interface {
	HttpHeader(ctx context.Context) (http.Header, error)
	HttpClient(ctx context.Context) (*http.Client, error)
	refresh(ctx context.Context) (*oauth2.Token, error)
}

type token struct {
	oauth       *oauth2.Token
	oauthConfig *oauth2.Config
}

func (t *token) refresh(ctx context.Context) (*oauth2.Token, error) {
	if t.oauth.Valid() {
		return t.oauth, nil
	}
	if t.oauthConfig == nil {
		return nil, fmt.Errorf("empty oauth config in token")
	}

	source := t.oauthConfig.TokenSource(ctx, t.oauth)
	newToken, err := source.Token()
	if err != nil {
		return nil, err
	}
	t.oauth = newToken
	return t.oauth, nil
}

func (t *token) HttpHeader(ctx context.Context) (http.Header, error) {
	oauth, err := t.refresh(ctx)
	if err != nil {
		return nil, err
	}
	return http.Header{
		"Authorization": []string{"Bearer " + oauth.AccessToken},
	}, nil
}

func (t *token) HttpClient(ctx context.Context) (*http.Client, error) {
	accessToken, err := t.refresh(ctx)
	if err != nil {
		return nil, err
	}
	source := t.oauthConfig.TokenSource(ctx, accessToken)
	return oauth2.NewClient(ctx, source), nil
}
