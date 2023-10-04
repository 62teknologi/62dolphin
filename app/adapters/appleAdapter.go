package adapters

import (
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
)

type AppleAdapter struct {
	config *oauth2.Config
}

func (adp *AppleAdapter) Init() interfaces.AuthInterface {
	// TODO handle Init method
	return adp
}

func (adp *AppleAdapter) GenerateLoginURL() string {
	return adp.config.AuthCodeURL("")
}

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
	// TODO handle Callback method

	return nil
}

func (adp *AppleAdapter) getProfile(ctx *gin.Context) (*Profile, error) {
	// TODO handle getProfile method

	return nil, nil
}

func (adp *AppleAdapter) generateToken(ctx *gin.Context, email string) (map[string]any, error) {
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
