package auth

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"net/http"
)

type PrivyOAuth struct {
	oauthConfig *oauth2.Config
}

func NewPrivyOAuth() *PrivyOAuth {
	return &PrivyOAuth{
		oauthConfig: &oauth2.Config{
			ClientID:     config.PrivyAuthClientId,
			ClientSecret: config.PrivyAuthClientSecret,
			RedirectURL:  config.PrivyAuthRedirectUrl,
			Scopes:       []string{"read", "write"},
			Endpoint: oauth2.Endpoint{
				AuthURL: config.PrivyAuthUrl,
			}},
	}
}

func (p *PrivyOAuth) Exchange(_ *gin.Context, code string) (*oauth2.Token, error) {
	// create request body as a JSON-encoded byte slice
	requestBody, err := json.Marshal(map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     config.PrivyAuthClientId,
		"client_secret": config.PrivyAuthClientSecret,
		"redirect_uri":  config.PrivyAuthRedirectUrl,
		"code":          code,
	})
	if err != nil {
		return nil, err
	}

	// create HTTP request with POST method and request body
	req, err := http.NewRequest("POST", config.PrivyAuthTokenExchangeUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	// set request headers
	req.Header.Set("Content-Type", "application/json")

	// send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read HTTP response body
	var response *oauth2.Token
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	// print HTTP response status code and body
	return response, nil
}

func (p *PrivyOAuth) GenerateLoginURL() string {
	url := p.oauthConfig.AuthCodeURL("")
	return url
}

func (p *PrivyOAuth) LoginCallback(ctx *gin.Context) (*oauth2.Token, error) {
	code := ctx.Query("code")
	token, err := p.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	return token, nil
}
