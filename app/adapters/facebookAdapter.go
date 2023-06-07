package adapters

import (
	"encoding/json"
	"fmt"
	"log"
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

	client := adp.config.Client(ctx, token)

	response, err := client.Get(facebook.Endpoint.TokenURL + "/me?fields=id,name,email")
	if err != nil {
		fmt.Println("Failed to fetch user profile:", err)
	}
	defer response.Body.Close()

	// Parse the user profile JSON response
	// Customize this based on your application's needs
	var profile struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	err = json.NewDecoder(response.Body).Decode(&profile)

	if err != nil {
		log.Println("Failed to decode user profile:", err)
	}

	fmt.Println(token)
	fmt.Println(profile)

	return token, nil
}
