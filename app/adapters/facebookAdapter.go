package adapters

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/62teknologi/62dolphin/62golib/utils"
	"github.com/62teknologi/62dolphin/app/config"
	"github.com/62teknologi/62dolphin/app/interfaces"
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
	return adp.config.AuthCodeURL("")
}

func (adp *FacebookAdapter) Callback(ctx *gin.Context) error {
	profile, err := adp.getProfile(ctx)
	utils.LogJson(profile)

	return err
}

func (adp *FacebookAdapter) getProfile(ctx *gin.Context) (*Profile, error) {
	code := ctx.Query("code")
	token, err := adp.config.Exchange(ctx, code)

	if err != nil {
		fmt.Println(http.StatusInternalServerError, utils.ResponseData("error", "Error getting token from Facebook", nil))
		return nil, err
	}

	client := adp.config.Client(ctx, token)
	response, err := client.Get("https://graph.facebook.com/me?fields=id,name,email,birthday,gender,picture&access_token=" + token.AccessToken)

	if err != nil {
		fmt.Println("Failed to fetch user profile:", err)
		return nil, err
	}

	defer response.Body.Close()

	var fProfile map[string]any
	err = json.NewDecoder(response.Body).Decode(&fProfile)

	Profile := Profile{}

	if fProfile["id"] != nil {
		Profile.Fbid = fProfile["id"].(string)
	}

	if fProfile["name"] != nil {
		Profile.Name = fProfile["name"].(string)
	}

	if fProfile["email"] != nil {
		Profile.Email = fProfile["email"].(string)
	}

	if fProfile["gender"] != nil {
		Profile.Gender = fProfile["gender"].(string)
	}

	if fProfile["birthday"] != nil {
		Profile.Birthdate = fProfile["birthday"].(string)
	}

	if picture, ok := fProfile["picture"].(map[string]any); ok {
		if data, ok := picture["data"].(map[string]any); ok {
			if url, ok := data["url"].(string); ok {
				Profile.Photo = url
			}
		}
	}

	if err != nil {
		fmt.Println("Failed to decode user profile:", err)
	}

	return &Profile, nil
}
