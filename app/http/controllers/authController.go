package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/62teknologi/62dolphin/62golib/utils"
	"github.com/62teknologi/62dolphin/app/adapters"
	"github.com/62teknologi/62dolphin/app/config"
	"github.com/62teknologi/62dolphin/app/interfaces"
	"github.com/62teknologi/62dolphin/app/tokens"
	dutils "github.com/62teknologi/62dolphin/app/utils"

	"github.com/dbssensei/ordentmarketplace/util"
	"github.com/gin-gonic/gin"
)

func Login(ctx *gin.Context) {
	var adapter interfaces.AuthInterface = adapters.GetAdapter(ctx.Param("adapter")).Init()
	ctx.Redirect(http.StatusTemporaryRedirect, adapter.GenerateLoginURL())
}

func Callback(ctx *gin.Context) {
	var adapter interfaces.AuthInterface = adapters.GetAdapter(ctx.Param("adapter")).Init()
	adapter.Callback(ctx)
}

func SignIn(ctx *gin.Context) {
	tokenMaker, err := tokens.NewJWTMaker(config.Data.TokenSymmetricKey)

	fmt.Println("ini")
	if err != nil {
		fmt.Errorf("cannot create token maker: %w", err)
		return
	}

	// build and cleansing login input from json file
	input, err := utils.JsonFileParser("transformers/request/auth/login.json")
	var requestBody map[string]any
	if err := ctx.BindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}
	utils.MapValuesShifter(input, requestBody)
	utils.MapNullValuesRemover(input)

	// Query database and additional query
	var user map[string]any

	fmt.Println(input["email"])
	utils.DB.Table("users").Where(utils.DB.Where("email = ?", input["email"])).Take(&user)

	if user["id"] == nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "invalid email or password", nil))
		return
	}

	err = util.CheckPassword(input["password"].(string), user["password"].(string))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "invalid email or password", nil))
		return
	}

	uId, _ := strconv.ParseInt(fmt.Sprintf("%v", user["id"]), 10, 32)
	accessToken, accessPayload, err := tokenMaker.CreateToken(
		int32(uId),
		config.Data.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}

	refreshToken, refreshPayload, err := tokenMaker.CreateToken(
		int32(uId),
		config.Data.RefreshTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}

	// Store sessions data to DB
	params := map[string]any{
		"id":            refreshPayload.Id,
		"user_id":       int32(uId),
		"refresh_token": refreshToken,
		"platform_id":   int32(input["platform_id"].(float64)),
		"is_blocked":    false,
		"expires_at":    refreshPayload.ExpiredAt,
		"created_at":    time.Now(),
		"updated_at":    time.Now(),
	}

	session := utils.DB.Table("tokens").Create(&params)
	if session.Error != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", fmt.Sprintf("%v", session.Error.Error()), nil))
	}

	// Setup output to client
	defaultResponse := map[string]any{
		"session_id":               params["id"],
		"access_token":             accessToken,
		"access_token_expires_at":  accessPayload.ExpiredAt,
		"refresh_token":            refreshToken,
		"refresh_token_expires_at": refreshPayload.ExpiredAt,
		"platform_id":              params["platform_id"],
	}

	customResponse, err := utils.JsonFileParser("transformers/response/auth/login.json")
	customUser := customResponse["user"]

	utils.MapValuesShifter(customResponse, defaultResponse)
	if customUser != nil {
		utils.MapValuesShifter(customUser.(map[string]any), user)
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "sign-in successfully", customResponse))
}

type emailForgotPasswordParams struct {
	URL     string
	Name    string
	Subject string
}

func ForgotPassword(ctx *gin.Context) {
	// Parse and cleaning input
	input, err := utils.JsonFileParser("transformers/request/auth/forgot_password.json")
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

	if input["method"] == "email" {
		// Generate and create OTP
		otpCode, _ := dutils.GenerateOTP(8)
		otpParams := map[string]any{
			"type":       "email",
			"code":       otpCode,
			"content":    input["receiver"],
			"expires_at": time.Now().Local().Add(time.Minute * 30),
			"created_at": time.Now(),
			"updated_at": time.Now(),
		}
		createOtp := utils.DB.Table("otps").Create(otpParams)
		if createOtp.Error != nil {
			ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", createOtp.Error.Error(), nil))
			return
		}

		// Setup email sender
		receiverList := []dutils.EmailReceiver{
			{
				Name:    input["receiver"].(string),
				Address: input["receiver"].(string),
			},
		}

		// Send email verification
		go func() {
			resetToken := dutils.Encode(fmt.Sprintf("email#%v#%v", otpCode, input["receiver"]))
			emailData := emailForgotPasswordParams{
				URL:     input["redirect_url"].(string) + resetToken,
				Name:    input["receiver"].(string),
				Subject: "Your password reset token (valid for 10min)",
			}
			dutils.EmailSender("forgot_password.html", emailData, receiverList)
		}()
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "send forgot password otp successfully", nil))
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
	decodePasswordToken, err := dutils.Decode(ctx.Param("token"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}
	splitToken := strings.Split(decodePasswordToken, "#")

	// Check if user exist in db
	var otp map[string]any
	utils.DB.Table("otps").Where("type = ?", splitToken[0]).Where("code = ?", splitToken[1]).Where("content = ?", splitToken[2]).Take(&otp)

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
