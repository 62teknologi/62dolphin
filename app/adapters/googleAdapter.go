package adapters

import (
	"fmt"
	"net/http"

	"github.com/62teknologi/62dolphin/app/config"
	"github.com/62teknologi/62dolphin/app/interfaces"
	"github.com/62teknologi/62dolphin/app/utils"
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
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "profile"},
		Endpoint:     google.Endpoint,
	}

	return adp
}

func (adp *GoogleAdapter) GenerateLoginURL() string {
	url := adp.config.AuthCodeURL("")
	return url
}

func (adp *GoogleAdapter) Callback(ctx *gin.Context) (*oauth2.Token, error) {
	code := ctx.Query("code")
	token, err := adp.config.Exchange(ctx, code)
	if err != nil {
		fmt.Println(http.StatusInternalServerError, utils.ResponseData("error", "Error getting token from Google", nil))
		return nil, err
	}

	profile, err := getProfileFromGoogle(ctx, token)

	if err != nil {
		fmt.Println(http.StatusInternalServerError, utils.ResponseData("error", "Error getting data user from Google", nil))
		return nil, err
	}

	var googleProfile interfaces.AdapterProfile

	if profile.Names != nil {
		googleProfile.Name = profile.Names[0].DisplayName
	}
	if profile.EmailAddresses != nil {
		googleProfile.Email = profile.EmailAddresses[0].Value
	}
	if profile.PhoneNumbers != nil {
		googleProfile.Phone = profile.PhoneNumbers[0].Value
	}
	if profile.Birthdays != nil {
		googleProfile.Birthday = profile.Birthdays[0].Text
	}
	if profile.Photos != nil {
		googleProfile.Photo = profile.Photos[0].Url
	}

	redirectUrl := fmt.Sprintf("%s/auth/google/callback?name=%s&email=%s&phone=%s&birthday=%s&photo=%s",
		config.Data.MonolithUrl+"/api/v1",
		googleProfile.Name,
		googleProfile.Email,
		googleProfile.Phone,
		googleProfile.Birthday,
		googleProfile.Photo,
	)

	fmt.Println(redirectUrl)
	fmt.Println(token)
	fmt.Println(profile)

	return token, err
}

func getProfileFromGoogle(ctx *gin.Context, token *oauth2.Token) (*people.Person, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	service, err := people.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	profile, err := service.People.Get("people/me").PersonFields("names,emailAddresses,phoneNumbers,birthdays,photos").Do()
	if err != nil {
		return nil, err
	}
	return profile, nil
}
