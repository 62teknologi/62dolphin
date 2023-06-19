package adapters

import (
	"fmt"
	"net/http"

	"github.com/62teknologi/62dolphin/62golib/utils"
	"github.com/62teknologi/62dolphin/app/config"
	"github.com/62teknologi/62dolphin/app/interfaces"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
)

type GoogleAdapter struct {
	config *oauth2.Config
}

func (adp *GoogleAdapter) Init() interfaces.AuthInterface {
	adp.config = &oauth2.Config{
		ClientID:     config.Data.GoogleAuthClientId,
		ClientSecret: config.Data.GoogleAuthClientSecret,
		RedirectURL:  config.Data.GoogleAuthRedirectUrl,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/user.birthday.read",
			"https://www.googleapis.com/auth/user.phonenumbers.read",
		},
		Endpoint: google.Endpoint,
	}

	return adp
}

func (adp *GoogleAdapter) GenerateLoginURL() string {
	return adp.config.AuthCodeURL("")
}

func (adp *GoogleAdapter) Callback(ctx *gin.Context) error {
	profile, err := adp.getProfile(ctx)
	utils.LogJson(profile)

	return err
}

func (adp *GoogleAdapter) getProfile(ctx *gin.Context) (*Profile, error) {
	code := ctx.Query("code")
	token, err := adp.config.Exchange(ctx, code)

	if err != nil {
		fmt.Println(http.StatusInternalServerError, utils.ResponseData("error", "Error getting token from Google", nil))
		return nil, err
	}

	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	service, err := people.NewService(ctx, option.WithHTTPClient(client))

	if err != nil {
		return nil, err
	}

	gProfile, err := service.People.Get("people/me").PersonFields("names,emailAddresses,phoneNumbers,photos,birthdays").Do()
	utils.LogJson(gProfile)

	if err != nil {
		fmt.Println(http.StatusInternalServerError, utils.ResponseData("error", "Error getting data user from Google", nil))
		return nil, err
	}

	Profile := Profile{}

	if gProfile.Names != nil {
		Profile.ID = gProfile.Names[0].Metadata.Source.Id
		Profile.Name = gProfile.Names[0].DisplayName
	}

	if gProfile.EmailAddresses != nil {
		Profile.Email = gProfile.EmailAddresses[0].Value
	}

	if gProfile.PhoneNumbers != nil {
		Profile.Phone = gProfile.PhoneNumbers[0].Value
	}

	if gProfile.Birthdays != nil {
		Profile.Birthdate = gProfile.Birthdays[0].Text
	}

	if gProfile.Photos != nil {
		Profile.Photo = gProfile.Photos[0].Url
	}

	if gProfile.Birthdays != nil {
		//get birthday
	}

	return &Profile, err
}
