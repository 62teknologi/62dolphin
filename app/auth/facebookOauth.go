package auth

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

type FacebookOAuth struct {
	oauthConfig *oauth2.Config
}

func NewFacebookOAuth() *FacebookOAuth {
	return &FacebookOAuth{
		oauthConfig: &oauth2.Config{
			ClientID:     config.FacebookAuthClientId,
			ClientSecret: config.FacebookAuthClientSecret,
			RedirectURL:  config.FacebookAuthRedirectUrl,
			Scopes:       []string{"email"},
			Endpoint:     facebook.Endpoint,
		},
	}
}

func (g *FacebookOAuth) GenerateLoginURL() string {
	url := g.oauthConfig.AuthCodeURL("")
	return url
}

func (g *FacebookOAuth) LoginCallback(ctx *gin.Context) (*oauth2.Token, error) {
	code := ctx.Query("code")
	token, err := g.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	return token, nil
}
