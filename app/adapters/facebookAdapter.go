package adapters

import (
	"fmt"
	"net/http"

	"github.com/62teknologi/62dolphin/app/config"
	"github.com/62teknologi/62dolphin/app/interfaces"
	"github.com/62teknologi/62dolphin/app/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

type FacebookAdapter struct {
	config *oauth2.Config
}

func (adp *FacebookAdapter) Init() interfaces.AuthInterface {
	adp.config = &oauth2.Config{
		ClientID:     config.Data.FacebookAuthClientId,
		ClientSecret: config.Data.FacebookAuthClientSecret,
		RedirectURL:  config.Data.FacebookAuthRedirectUrl,
		Scopes:       []string{"email"},
		Endpoint:     facebook.Endpoint,
	}

	return adp
}

func (adp *FacebookAdapter) GenerateLoginURL() string {
	url := adp.config.AuthCodeURL("")
	return url
}

func (adp *FacebookAdapter) Callback(ctx *gin.Context) (*oauth2.Token, error) {
	code := ctx.Query("code")
	token, err := adp.config.Exchange(ctx, code)

	if err != nil {
		fmt.Println(http.StatusInternalServerError, utils.ResponseData("error", "Error getting token from Facebook", nil))
		return nil, err
	}

	/*
		TODO : Do something with the token, such as getting the user's email address
			using the Facebook API and redirect to client register page
	*/
	fmt.Println("Facebook token: %+v", token)

	return token, nil
}
