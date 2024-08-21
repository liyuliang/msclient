package msclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func TokenFromFile(path string) (Token, error) {
	if path == "" {
		return nil, fmt.Errorf("missing token file path")
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tmp := &tokenTpl{}
	if err = json.Unmarshal(b, tmp); err != nil {
		return nil, err
	}
	return &token{oauth: tmp.OAuth, oauthConfig: tmp.OAuthConfig}, nil
}
