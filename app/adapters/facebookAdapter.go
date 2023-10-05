package adapters

import (
	"encoding/json"
	"fmt"
	"github.com/62teknologi/62dolphin/app/tokens"
	"net/http"
	"strconv"
	"time"

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

func (adp *FacebookAdapter) Verify(ctx *gin.Context, email, userId string) (map[string]any, error) {
	var user map[string]any
	utils.DB.Table("users").Where("email = ?", email).Where("facebook_id = ?", userId).Take(&user)

	if user["id"] == nil {
		return map[string]any{
			"email":   email,
			"user_id": userId,
		}, fmt.Errorf("user not found")
	}

	token, err := adp.generateToken(ctx, email)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (adp *FacebookAdapter) Callback(ctx *gin.Context) error {
	profile, err := adp.getProfile(ctx)
	if err != nil {
		return fmt.Errorf("error while get profile")
	}

	token, err := adp.generateToken(ctx, profile.Email)
	if err != nil {
		return err
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
		Profile.ID = fProfile["id"].(string)
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

func (adp *FacebookAdapter) generateToken(ctx *gin.Context, email string) (map[string]any, error) {
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

	defaultResponse := map[string]any{
		"session_id":               params["id"],
		"access_token":             accessToken,
		"access_token_expires_at":  accessPayload.ExpiredAt,
		"refresh_token":            refreshToken,
		"refresh_token_expires_at": refreshPayload.ExpiredAt,
		"platform_id":              params["platform_id"],
	}

	customResponse, err := utils.JsonFileParser(config.Data.SettingPath + "/transformers/response/auth/login.json")
	customUser := customResponse["user"]

	utils.MapValuesShifter(customResponse, defaultResponse)

	if customUser != nil {
		utils.MapValuesShifter(customUser.(map[string]any), user)
	}

	return customResponse, nil
}
