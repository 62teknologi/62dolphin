package controllers

import (
	"github.com/62teknologi/62dolphin/62golib/utils"
	dutils "github.com/62teknologi/62dolphin/app/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type createOTPRequest struct {
	Method            string `json:"method" binding:"required"`
	Receiver          string `json:"receiver" binding:"required"`
	OTPLength         int    `json:"otp_length" binding:"required"`
	OTPExpiredMinutes int    `json:"otp_expired_minutes"`
}

func CreateOTP(ctx *gin.Context) { // Setup request body
	var req createOTPRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	var otpExpiredMinutes int
	if req.OTPExpiredMinutes == 0 {
		otpExpiredMinutes = 30
	} else {
		otpExpiredMinutes = req.OTPExpiredMinutes
	}

	otpCode, _ := dutils.GenerateOTP(req.OTPLength)

	otpParams := map[string]any{
		"type":       req.Method,
		"code":       otpCode,
		"receiver":   req.Receiver,
		"expires_at": time.Now().Local().Add(time.Minute * time.Duration(otpExpiredMinutes)),
		"created_at": time.Now(),
		"updated_at": time.Now()}

	var existingOtp map[string]any
	utils.DB.Table("otps").Where("receiver = ?", otpParams["receiver"]).Take(&existingOtp)

	if existingOtp["id"] != nil {
		updatedOtp := utils.DB.Table("otps").Where("receiver = ?", otpParams["receiver"]).Updates(otpParams)
		if updatedOtp.Error != nil {
			ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", updatedOtp.Error.Error(), nil))
			return
		}
	} else {
		createdOtp := utils.DB.Table("otps").Create(otpParams)
		if createdOtp.Error != nil {
			ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", createdOtp.Error.Error(), nil))
			return
		}

	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "success create otp", map[string]any{
		"otp_code": otpCode,
	}))
}
