package auth

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

type MicrosoftOAuth struct {
	oauthConfig *oauth2.Config
}

func NewMicrosoftOAuth() *MicrosoftOAuth {
	return &MicrosoftOAuth{
		oauthConfig: &oauth2.Config{
			ClientID:     config.MicrosoftAuthClientId,
			ClientSecret: config.MicrosoftAuthClientSecret,
			RedirectURL:  config.MicrosoftAuthRedirectUrl,
			Scopes:       []string{"email"},
			Endpoint:     microsoft.AzureADEndpoint(config.MicrosoftAuthTenantId),
		},
	}
}

func (g *MicrosoftOAuth) GenerateLoginURL() string {
	url := g.oauthConfig.AuthCodeURL("")
	return url
}

func (g *MicrosoftOAuth) LoginCallback(ctx *gin.Context) (*oauth2.Token, error) {
	code := ctx.Query("code")
	token, err := g.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	return token, nil
}
