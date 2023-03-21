package auth

import (
	"dolphin/app/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleOAuth struct {
	oauthConfig *oauth2.Config
}

var config, _ = utils.LoadConfig(".")

func NewGoogleOAuth() *GoogleOAuth {
	return &GoogleOAuth{
		oauthConfig: &oauth2.Config{
			ClientID:     config.GoogleAuthClientId,
			ClientSecret: config.GoogleAuthClientSecret,
			RedirectURL:  config.GoogleAuthRedirectUrl,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (g *GoogleOAuth) GenerateLoginURL() string {
	url := g.oauthConfig.AuthCodeURL("")
	return url
}

func (g *GoogleOAuth) LoginCallback(ctx *gin.Context) (*oauth2.Token, error) {
	code := ctx.Query("code")
	token, err := g.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	return token, nil
}
