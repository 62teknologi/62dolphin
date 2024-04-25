package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/62teknologi/62dolphin/62golib/utils"
	"github.com/62teknologi/62dolphin/app/config"
	"github.com/62teknologi/62dolphin/app/tokens"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

type accessTokenVerifyRequest struct {
	AccessToken string `json:"access_token" binding:"required"`
}

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
)

func VerifyAccessToken(ctx *gin.Context) {
	// Setup request body
	var req accessTokenVerifyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.ResponseData("error", err.Error(), nil))
		return
	}

	// Setup and check given token
	tokenMaker, err := tokens.NewJWTMaker(config.Data.TokenSymmetricKey)
	if err != nil {
		fmt.Errorf("cannot create token maker: %w", err)
		return
	}

	payload, err := tokenMaker.VerifyToken(req.AccessToken)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.ResponseData("error", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "verify token successfully", payload))
}

type createAccessTokenRequest struct {
	UserId int32 `json:"user_id" binding:"required"`
}

func CreateAccessToken(ctx *gin.Context) {
	// Setup request body
	var req createAccessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	tokenMaker, err := tokens.NewJWTMaker(config.Data.TokenSymmetricKey)
	if err != nil {
		fmt.Errorf("cannot create token maker: %w", err)
		return
	}

	accessToken, accessPayload, err := tokenMaker.CreateToken(
		req.UserId,
		config.Data.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}

	refreshToken, refreshPayload, err := tokenMaker.CreateToken(
		req.UserId,
		config.Data.RefreshTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}

	// Store sessions data to DB
	params := map[string]any{
		"id":            refreshPayload.Id,
		"user_id":       req.UserId,
		"refresh_token": refreshToken,
		"is_blocked":    false,
		"platform_id":   0,
		"expires_at":    refreshPayload.ExpiredAt,
		"created_at":    time.Now(),
		"updated_at":    time.Now(),
	}

	session := utils.DB.Table("tokens").Create(&params)
	if session.Error != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", fmt.Sprintf("%v", session.Error.Error()), nil))
		return
	}

	// Setup output to client
	res := map[string]any{
		"session_id":               params["id"],
		"access_token":             accessToken,
		"access_token_expires_at":  accessPayload.ExpiredAt,
		"refresh_token":            refreshToken,
		"refresh_token_expires_at": refreshPayload.ExpiredAt,
		"platform_id":              params["platform_id"],
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "create token successfully", res))
}

type accessTokenRequest struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IsAllDevice  bool   `json:"is_all_device"`
}

type accessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func RenewAccessToken(ctx *gin.Context) {
	// Setup request body
	var req accessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	// Setup and check given token
	tokenMaker, err := tokens.NewJWTMaker(config.Data.TokenSymmetricKey)
	if err != nil {
		fmt.Errorf("cannot create token maker: %w", err)
		return
	}
	refreshPayload, err := tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.ResponseData("error", err.Error(), nil))
		return
	}

	// Get refresh token data from database
	var token map[string]any
	tokenQuery := utils.DB.Table("tokens").Where("refresh_token = ?", req.RefreshToken).Take(&token)

	// Check token validity
	if tokenQuery.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(tokenQuery.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			ctx.JSON(http.StatusUnauthorized, utils.ResponseData("error", tokenQuery.Error.Error(), nil))
			return
		}
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", tokenQuery.Error.Error(), nil))
		return
	}

	if utils.ConvertToInt(token["is_blocked"]) == 1 {
		ctx.JSON(http.StatusUnauthorized, utils.ResponseData("error", "token has been blocked", nil))
		return
	}

	if int32(token["user_id"].(int64)) != refreshPayload.UserId {
		ctx.JSON(http.StatusUnauthorized, utils.ResponseData("error", "incorrect token user", nil))
		return
	}
	if token["refresh_token"] != req.RefreshToken {
		ctx.JSON(http.StatusUnauthorized, utils.ResponseData("error", "mismatched token token", nil))
		return
	}
	if time.Now().After(token["expires_at"].(time.Time)) {
		ctx.JSON(http.StatusUnauthorized, utils.ResponseData("error", "expired token", nil))
		return
	}

	// Generate new access token
	accessToken, accessPayload, err := tokenMaker.CreateToken(
		refreshPayload.UserId,
		config.Data.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}

	// Setup and send response
	rsp := accessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "refresh token successfully", rsp))
}

func BlockRefreshToken(ctx *gin.Context) {
	// Setup request body
	var req accessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	// Get token auth payload
	authorizationPayload, _ := ctx.Get("authorization_payload")

	// Update blocked token on db
	tokenQuery := utils.DB.Table("tokens").
		Where("user_id", authorizationPayload.(*tokens.Payload).UserId).
		Where("refresh_token = ?", req.RefreshToken).
		Update("is_blocked", true)

	// Handle query error
	if tokenQuery.Error != nil {
		if tokenQuery.Error.Error() == "record not found" {
			ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "token data not found", nil))
			return
		}
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", tokenQuery.Error.Error(), nil))
		return
	}

	// Send success response to client
	ctx.JSON(http.StatusOK, utils.ResponseData("success", "blocking token successfully", nil))
}

func BlockAllRefreshToken(ctx *gin.Context) {
	// Get token auth payload
	authorizationPayload, _ := ctx.Get("authorization_payload")

	// Update blocked token on db
	tokenQuery := utils.DB.Table("tokens").
		Where("user_id", authorizationPayload.(*tokens.Payload).UserId).
		Update("is_blocked", true)

	// Handle query error
	if tokenQuery.Error != nil {
		if tokenQuery.Error.Error() == "record not found" {
			ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "token data not found", nil))
			return
		}
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", tokenQuery.Error.Error(), nil))
		return
	}

	// Send success response to client
	ctx.JSON(http.StatusOK, utils.ResponseData("success", "blocking token successfully", nil))
}

func RevokeRefreshToken(ctx *gin.Context) {
	// Setup request body
	var req accessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	// Get & check refresh token data from database
	var token map[string]any
	getTokenQuery := utils.DB.Table("tokens").Where("refresh_token = ?", req.RefreshToken).Take(&token)

	// Handle query error
	if getTokenQuery.Error != nil {
		if getTokenQuery.Error.Error() == "record not found" {
			ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "token data not found", nil))
			return
		}
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", getTokenQuery.Error.Error(), nil))
		return
	}

	// Get token auth payload
	authorizationPayload, _ := ctx.Get("authorization_payload")

	// Delete access token on db
	tokenQuery := utils.DB.Table("tokens").
		Where("user_id", authorizationPayload.(*tokens.Payload).UserId)

	if req.IsAllDevice == true {
		tokenQuery = tokenQuery.Delete(nil)
	} else {
		tokenQuery = tokenQuery.Where("refresh_token = ?", req.RefreshToken).Delete(nil)
	}

	// Handle query error
	if tokenQuery.Error != nil {
		if tokenQuery.Error.Error() == "record not found" {
			ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "token data not found", nil))
			return
		}
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", tokenQuery.Error.Error(), nil))
		return
	}

	// Send success response to client
	ctx.JSON(http.StatusOK, utils.ResponseData("success", "revoke token successfully", nil))
}

func RevokeAllRefreshToken(ctx *gin.Context) {
	// Get token auth payload
	authorizationPayload, _ := ctx.Get("authorization_payload")

	// Update blocked token on db
	tokenQuery := utils.DB.Table("tokens").
		Where("user_id", authorizationPayload.(*tokens.Payload).UserId).Delete(nil)

	if tokenQuery.RowsAffected <= 0 {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "token data not found", nil))
		return
	}

	// Handle query error
	if tokenQuery.Error != nil {
		if tokenQuery.Error.Error() == "record not found" {
			ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "token data not found", nil))
			return
		}
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", tokenQuery.Error.Error(), nil))
		return
	}

	// Send success response to client
	ctx.JSON(http.StatusOK, utils.ResponseData("success", "successfully revoke all token", nil))
}

func RevokeAccessToken(ctx *gin.Context) {
	// Setup request body
	var req accessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	// Get access token
	authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
	if len(authorizationHeader) == 0 {
		err := errors.New("authorization header is not provided")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.ResponseData("error", err.Error(), nil))

		return
	}

	fields := strings.Fields(authorizationHeader)

	if len(fields) < 2 {
		err := errors.New("invalid authorization header format")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.ResponseData("error", err.Error(), nil))

		return
	}

	authorizationType := strings.ToLower(fields[0])

	if authorizationType != authorizationTypeBearer {
		err := fmt.Errorf("unsupported authorization type %s", authorizationType)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, utils.ResponseData("error", err.Error(), nil))

		return
	}

	accessToken := fields[1]

	// Get token auth payload
	authorizationPayload, _ := ctx.Get("authorization_payload")

	// Update blocked token on db
	tokenQuery := utils.DB.Table("tokens").
		Where("user_id", authorizationPayload.(*tokens.Payload).UserId)

	if req.IsAllDevice == true {
		tokenQuery = tokenQuery.Update("is_blocked", true)
	} else {
		tokenQuery = tokenQuery.Where("access_token = ?", accessToken).Update("is_blocked", true)
	}

	if tokenQuery.RowsAffected <= 0 {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "token data not found", nil))
		return
	}

	// Handle query error
	if tokenQuery.Error != nil {
		if tokenQuery.Error.Error() == "record not found" {
			ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "token data not found", nil))
			return
		}
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", tokenQuery.Error.Error(), nil))
		return
	}

	// Send success response to client
	ctx.JSON(http.StatusOK, utils.ResponseData("success", "blocking token successfully", nil))
}
