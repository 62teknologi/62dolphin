package adapters

import (
	"fmt"
	"github.com/62teknologi/62dolphin/app/tokens"
	"net/http"
	"strconv"
	"time"
	"encoding/json"
	"github.com/62teknologi/62dolphin/62golib/utils"
	"github.com/62teknologi/62dolphin/app/config"
	"github.com/62teknologi/62dolphin/app/interfaces"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type AppleAdapter struct {
	config *oauth2.Config
}

// Init initializes the AppleAdapter with OAuth2 configuration
func (adp *AppleAdapter) Init() interfaces.AuthInterface {
	adp.config = &oauth2.Config{
		ClientID:     config.Data.AppleClientID,
		ClientSecret: config.Data.AppleClientSecret,
		RedirectURL:  config.Data.AppleRedirectURL,
		Scopes:       []string{"email", "name"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://appleid.apple.com/auth/authorize",
			TokenURL: "https://appleid.apple.com/auth/token",
		},
	}
	return adp
}

// GenerateLoginURL generates the login URL to start the OAuth2 authentication process
func (adp *AppleAdapter) GenerateLoginURL() string {
	return adp.config.AuthCodeURL("")
}

// Verify verifies the user based on email and userId
func (adp *AppleAdapter) Verify(ctx *gin.Context, email, userId string) (map[string]any, error) {
	var user map[string]any
	utils.DB.Table("users").Where("email = ?", email).Where("apple_id = ?", userId).Take(&user)

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


func (adp *AppleAdapter) Callback(ctx *gin.Context) error {
	code := ctx.DefaultQuery("code", "")
	if code == "" {
			return fmt.Errorf("missing authorization code")
	}

	// Exchange the authorization code for a token
	token, err := adp.config.Exchange(ctx, code)
	if err != nil {
			return fmt.Errorf("failed to exchange authorization code for token: %v", err)
	}

	// Fetch the user's profile using the token
	profile, err := adp.getProfile(ctx, token)
	if err != nil {
			return fmt.Errorf("failed to get user profile: %v", err)
	}
	
	// Call the Verify method to create the session token
	_, err = adp.Verify(ctx, profile.Email, profile.ID)
	if err != nil {
			return fmt.Errorf("failed to verify user: %v", err)
	}

	return nil
}

// getProfile fetches the user profile from Apple's userinfo endpoint
func (adp *AppleAdapter) getProfile(ctx *gin.Context, token *oauth2.Token) (*Profile, error) {
	client := adp.config.Client(ctx, token)
	resp, err := client.Get("https://appleid.apple.com/auth/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user profile: %v", err)
	}
	defer resp.Body.Close()

	var profile Profile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("failed to parse user profile: %v", err)
	}
	return &profile, nil
}

func (adp *AppleAdapter) generateToken(ctx *gin.Context, email string) (map[string]any, error) {
	tokenMaker, err := tokens.NewJWTMaker(config.Data.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create token maker: %v", err)
	}

	var user map[string]any
	utils.DB.Table("users").Where("email = ?", email).Take(&user)

	uId, _ := strconv.ParseInt(fmt.Sprintf("%v", user["id"]), 10, 32)

	// Create Access Token
	accessToken, accessPayload, err := tokenMaker.CreateToken(
		int32(uId),
		config.Data.AccessTokenDuration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %v", err)
	}

	// Create Refresh Token
	refreshToken, refreshPayload, err := tokenMaker.CreateToken(
		int32(uId),
		config.Data.RefreshTokenDuration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %v", err)
	}

	// Store session data in DB
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

	if config.Data.TokenDestroy {
		params["access_token"] = accessToken
	}

	if err := utils.DB.Table("tokens").Create(&params).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", fmt.Sprintf("%v", err.Error()), nil))
	}

	defaultResponse := map[string]any{
		"session_id":               params["id"],
		"access_token":             accessToken,
		"access_token_expires_at": accessPayload.ExpiredAt,
		"refresh_token":            refreshToken,
		"refresh_token_expires_at": refreshPayload.ExpiredAt,
		"platform_id":              params["platform_id"],
	}

	// Customize response
	customResponse, err := utils.JsonFileParser(config.Data.SettingPath + "/transformers/response/auth/login.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load custom response: %v", err)
	}

	customUser := customResponse["user"]
	utils.MapValuesShifter(customResponse, defaultResponse)

	if customUser != nil {
		utils.MapValuesShifter(customUser.(map[string]any), user)
	}

	return customResponse, nil
}
