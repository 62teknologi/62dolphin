package adapters

import (
	"fmt"
	"github.com/62teknologi/62dolphin/app/tokens"
	"github.com/goccy/go-json"
	"io"
	"net/http"
	"strconv"
	"time"

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
	redirectUrl := fmt.Sprintf("%s/auth/facebook/callback?token=%v&auth-token=%v", config.Data.MonolithUrl+"/api/v1", encodedProfile, encodeToken)

	ctx.Header("Authorization", "Basic "+utils.Encode(config.Data.ApiKey))
	ctx.Redirect(http.StatusTemporaryRedirect, redirectUrl)

	return nil
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

func (adp *MicrosoftAdapter) generateToken(ctx *gin.Context, email string) (map[string]any, error) {
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
