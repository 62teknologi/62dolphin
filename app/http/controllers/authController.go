package controllers

import (
	"fmt"
	dutils "github.com/62teknologi/62dolphin/app/utils"
	"net/http"
	"strings"
	"time"

	"github.com/62teknologi/62dolphin/62golib/utils"
	"github.com/62teknologi/62dolphin/app/adapters"
	"github.com/62teknologi/62dolphin/app/config"

	"github.com/dbssensei/ordentmarketplace/util"
	"github.com/gin-gonic/gin"
)

func Login(ctx *gin.Context) {
	adapter, err := adapters.GetAdapter(ctx.Param("adapter"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("success", err.Error(), nil))
	}

	adapter = adapter.Init()
	ctx.Redirect(http.StatusTemporaryRedirect, adapter.GenerateLoginURL())
}

func Callback(ctx *gin.Context) {
	adapterName := ctx.Param("adapter")

	if strings.Contains(ctx.FullPath(), "/auth/sign-in") {
		adapterName = "local"
	}

	adapter, err := adapters.GetAdapter(adapterName)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("success", err.Error(), nil))
	}

	adapter = adapter.Init()
	adapter.Callback(ctx)
}

func ForgotPassword(ctx *gin.Context) {
	// Parse and cleaning input
	input, err := utils.JsonFileParser(config.Data.SettingPath + "/transformers/request/auth/forgot_password.json")
	var userInput map[string]any
	if err = ctx.BindJSON(&userInput); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}
	utils.MapValuesShifter(input, userInput)
	utils.MapNullValuesRemover(input)

	// Check if user exist in db
	var user map[string]any
	utils.DB.Table("users").Where(input["method"].(string)+" = ?", input["receiver"]).Take(&user)

	if user["id"] == nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "invalid user", nil))
		return
	}

	otpCode, _ := dutils.GenerateOTP(8)
	otpParams := map[string]any{
		"type":       "email",
		"code":       otpCode,
		"receiver":   input["receiver"],
		"expires_at": time.Now().Local().Add(time.Minute * 30),
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
	createOtp := utils.DB.Table("otps").Create(otpParams)
	if createOtp.Error != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", createOtp.Error.Error(), nil))
		return
	}
	resetToken := utils.Encode(fmt.Sprintf("%v#%v#%v", input["method"], otpCode, input["receiver"]))

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "send forgot password otp successfully", map[string]any{
		"reset_token": resetToken,
	}))
}

type resetPasswordParams struct {
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

func ResetPassword(ctx *gin.Context) {
	var req resetPasswordParams
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	// decodePasswordToken should return <otp type>#<otp code>#<otp content>
	decodePasswordToken, err := utils.Decode(ctx.Param("token"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}
	splitToken := strings.Split(decodePasswordToken, "#")

	// Check if user exist in db
	var otp map[string]any
	utils.DB.Table("otps").Where("type = ?", splitToken[0]).Where("code = ?", splitToken[1]).Where("receiver = ?", splitToken[2]).Take(&otp)

	if otp["id"] == nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "invalid otp", nil))
		return
	}

	if otp["expires_at"].(time.Time).Unix() < time.Now().Unix() {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "otp expired", nil))
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "new password is not the same as confirm password", nil))
		return
	}

	// Hashing Password
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", fmt.Sprintf("%v", err.Error()), nil))
		return
	}

	var user map[string]any
	utils.DB.Table("users").Where(splitToken[0]+" = ?", splitToken[2]).Update("password", hashedPassword)
	fmt.Println(user)

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "reset password successfully", nil))
}
