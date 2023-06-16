package adapters

import (
	"fmt"
	"net/http"

	"github.com/62teknologi/62dolphin/62golib/utils"
	"github.com/62teknologi/62dolphin/app/config"
	"github.com/62teknologi/62dolphin/app/interfaces"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

type MicrosoftAdapter struct {
	config *oauth2.Config
}

func (adp *MicrosoftAdapter) Init() interfaces.AuthInterface {
	adp.config = &oauth2.Config{
		ClientID:     config.Data.MicrosoftAuthClientId,
		ClientSecret: config.Data.MicrosoftAuthClientSecret,
		RedirectURL:  config.Data.MicrosoftAuthRedirectUrl,
		Scopes:       []string{"email"},
		Endpoint:     microsoft.AzureADEndpoint(config.Data.MicrosoftAuthTenantId),
	}

	return adp
}

func (adp *MicrosoftAdapter) GenerateLoginURL() string {
	url := adp.config.AuthCodeURL("")
	return url
}

func (adp *MicrosoftAdapter) Callback(ctx *gin.Context) error {
	code := ctx.Query("code")
	token, err := adp.config.Exchange(ctx, code)
	if err != nil {
		fmt.Println(http.StatusInternalServerError, utils.ResponseData("error", "Error getting token from Microsoft", nil))
		return err
	}

	// Use the token to access the user's profile.
	/*
		TODO : Do something with the token, such as getting the user's email address
			using the Microsoft API and redirect to client register page
	*/
	fmt.Println("Microsoft token: %+v", token)

	return nil
}
