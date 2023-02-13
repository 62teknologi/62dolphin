package controllers

import (
	"dolphin/app/models"
	"dolphin/app/tokens"
	"dolphin/app/utils"
	"fmt"
	"github.com/dbssensei/ordentmarketplace/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type loginUserResponse struct {
	SessionID             uuid.UUID   `json:"session_id"`
	AccessToken           string      `json:"access_token"`
	AccessTokenExpiresAt  time.Time   `json:"access_token_expires_at"`
	RefreshToken          string      `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time   `json:"refresh_token_expires_at"`
	PlatformId            int32       `json:"platform_id"`
	User                  models.User `json:"user"`
}

func SignIn(ctx *gin.Context) {
	config, err := utils.LoadConfig(".")
	if err != nil {
		fmt.Errorf("cannot load config: %w", err)
		return
	}

	tokenMaker, err := tokens.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		fmt.Errorf("cannot create token maker: %w", err)
		return
	}

	// build and cleansing login input from json file
	jsonInput, err := utils.JsonFileParser("transformers/auth/loginInput.json")
	var userInput map[string]any
	if err := ctx.BindJSON(&userInput); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}
	utils.MapValuesShifter(jsonInput, userInput)
	utils.MapNullValuesRemover(jsonInput)

	// Query database and additional query
	var user models.User
	query := models.DB.Where("is_active = true").Where("email = ?", jsonInput["email"]).Or("username = ?", jsonInput["username"])
	query.First(&user)

	// Validate user and comparing password
	if user.Id == 0 {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "invalid email or password", nil))
		return
	}

	err = util.CheckPassword(jsonInput["password"].(string), user.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "invalid email or password", nil))
		return
	}

	accessToken, accessPayload, err := tokenMaker.CreateToken(
		user.Id,
		config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}

	refreshToken, refreshPayload, err := tokenMaker.CreateToken(
		user.Id,
		config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}

	// Store sessions data to DB
	arg := models.Token{
		Id:           refreshPayload.Id,
		UserId:       user.Id,
		RefreshToken: refreshToken,
		PlatformId:   int32(jsonInput["platform_id"].(float64)),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	session := models.DB.Create(&arg)
	if session.Error != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", fmt.Sprintf("%v", err.Error()), nil))
	}

	// Setup output to user
	rsp := loginUserResponse{
		SessionID:             arg.Id,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		PlatformId:            arg.PlatformId,
		User:                  user,
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "sign-in successfully", rsp))
}
func SignUp(c *gin.Context) {}
