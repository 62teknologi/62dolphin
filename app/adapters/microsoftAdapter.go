package adapters

import (
	"fmt"
	"github.com/goccy/go-json"
	"io"
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
	microsoftEndpoint := microsoft.AzureADEndpoint(config.Data.MicrosoftAuthTenantId)
	adp.config = &oauth2.Config{
		ClientID:     config.Data.MicrosoftAuthClientId,
		ClientSecret: config.Data.MicrosoftAuthClientSecret,
		RedirectURL:  config.Data.MicrosoftAuthRedirectUrl,
		Scopes:       []string{"User.Read"},
		Endpoint:     microsoftEndpoint,
	}

	return adp
}

func (adp *MicrosoftAdapter) GenerateLoginURL() string {
	url := adp.config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	return url
}

func (adp *MicrosoftAdapter) Callback(ctx *gin.Context) error {
	profile, err := adp.getUserProfile(ctx)
	utils.LogJson(profile)

	return err
}

type UserProfile struct {
	ID    string `json:"id"`
	Name  string `json:"displayName"`
	Email string `json:"mail"`
	Phone string `json:"mobilePhone"`
}

func (adp *MicrosoftAdapter) getUserProfile(ctx *gin.Context) (*Profile, error) {
	code := ctx.Query("code")
	token, err := adp.config.Exchange(ctx, code)

	if err != nil {
		fmt.Println(http.StatusInternalServerError, utils.ResponseData("error", "Error getting token from Facebook", nil))
		return nil, err
	}

	client := adp.config.Client(ctx, token)
	response, err := client.Get("https://graph.microsoft.com/v1.0/me")

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var userProfile UserProfile
	err = json.Unmarshal(body, &userProfile)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	profile := &Profile{
		ID:    userProfile.ID,
		Name:  userProfile.Name,
		Email: userProfile.Email,
		Phone: userProfile.Phone,
		//Photo:     "", // Get profile image has difference url
		//Gender:    "",
		//Birthdate: "",
		//AgeMin:    0,
		//AgeMax:    0,
		//AgeRange:  "",
	}

	return profile, nil
}
