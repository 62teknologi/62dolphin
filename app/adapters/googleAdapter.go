package adapters

import (
	"fmt"
	"github.com/62teknologi/62dolphin/app/tokens"
	"github.com/goccy/go-json"
	"net/http"
	"strconv"
	"time"

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
	if err != nil {
		return fmt.Errorf("error while get profile")
	}

	token, err := adp.generateToken(ctx, profile.Email)
	if err != nil {
		return fmt.Errorf("error while get profile")
	}

	profileJson, _ := json.Marshal(profile)
	encodedProfile := utils.Encode(string(profileJson))
	tokenJson, _ := json.Marshal(token)
	encodeToken := utils.Encode(string(tokenJson))
	redirectUrl := fmt.Sprintf("%s/auth/google/callback?token=%v&auth-token=%v", config.Data.MonolithUrl+"/api/v1", encodedProfile, encodeToken)

	ctx.Header("Authorization", "Basic "+utils.Encode(config.Data.ApiKey))
	ctx.Redirect(http.StatusTemporaryRedirect, redirectUrl)

	return nil
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

func (adp *GoogleAdapter) generateToken(ctx *gin.Context, email string) (map[string]any, error) {
	tokenMaker, err := tokens.NewJWTMaker(config.Data.TokenSymmetricKey)

	var user map[string]any
	utils.DB.Table("users").Where("email = ?", email).Take(&user)

	uId, _ := strconv.ParseInt(fmt.Sprintf("%v", user["id"]), 10, 32)

	accessToken, accessPayload, err := tokenMaker.CreateToken(
		int32(uId),
		config.Data.AccessTokenDuration,
	)

	if err != nil {
		return nil, err
	}

	refreshToken, refreshPayload, err := tokenMaker.CreateToken(
		int32(uId),
		config.Data.RefreshTokenDuration,
	)

	if err != nil {
		return nil, err
	}

	// Store sessions data to DB
	params := map[string]any{
		"id":            refreshPayload.Id,
		"user_id":       int32(uId),
		"refresh_token": refreshToken,
		"platform_id":   1,
		"is_blocked":    false,
		"expires_at":    refreshPayload.ExpiredAt,
		"created_at":    time.Now(),
		"updated_at":    time.Now(),
	}

	if err := utils.DB.Table("tokens").Create(&params).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", fmt.Sprintf("%v", err.Error()), nil))
	}

	return map[string]any{
		"session_id":               params["id"],
		"access_token":             accessToken,
		"access_token_expires_at":  accessPayload.ExpiredAt,
		"refresh_token":            refreshToken,
		"refresh_token_expires_at": refreshPayload.ExpiredAt,
		"platform_id":              params["platform_id"],
	}, nil
}
