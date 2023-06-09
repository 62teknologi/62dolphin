package controllers

import (
	"github.com/62teknologi/62dolphin/62golib/utils"
	dutils "github.com/62teknologi/62dolphin/app/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type createOTPRequest struct {
	Method    string `json:"method" binding:"required"`
	Receiver  string `json:"receiver" binding:"required"`
	OTPLength int    `json:"otp_length" binding:"required"`
}

func CreateOTP(ctx *gin.Context) { // Setup request body
	var req createOTPRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	otpCode, _ := dutils.GenerateOTP(req.OTPLength)

	otpParams := map[string]any{
		"type":       req.Method,
		"code":       otpCode,
		"receiver":   req.Receiver,
		"expires_at": time.Now().Local().Add(time.Minute * 30),
		"created_at": time.Now(),
		"updated_at": time.Now()}

	createOtp := utils.DB.Table("otps").Create(otpParams)

	if createOtp.Error != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", createOtp.Error.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "success create otp", map[string]any{
		"otp_code": otpCode,
	}))
}
