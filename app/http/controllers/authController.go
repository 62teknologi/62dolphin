package controllers

import (
	"dolphin/app/auth"
	"dolphin/app/tokens"
	"dolphin/app/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dbssensei/ordentmarketplace/util"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
)

var googleAdapter = auth.NewGoogleOAuth()

func GoogleLogin(ctx *gin.Context) {
	ctx.Redirect(http.StatusTemporaryRedirect, googleAdapter.GenerateLoginURL())
}

func GoogleCallback(ctx *gin.Context) {
	config, err := utils.LoadConfig(".")
	if err != nil {
		fmt.Errorf("cannot load config: %w", err)
		return
	}

	token, err := googleAdapter.LoginCallback(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", "Error getting token from Google", nil))
		return
	}

	profile, err := getProfileFromGoogle(ctx, token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", "Error getting data user from Google", nil))
		return
	}

	var googleProfile auth.OAuthProfile

	if profile.Names != nil {
		googleProfile.Name = profile.Names[0].DisplayName
	}
	if profile.EmailAddresses != nil {
		googleProfile.Email = profile.EmailAddresses[0].Value
	}
	if profile.PhoneNumbers != nil {
		googleProfile.Phone = profile.PhoneNumbers[0].Value
	}
	if profile.Birthdays != nil {
		googleProfile.Birthday = profile.Birthdays[0].Text
	}
	if profile.Photos != nil {
		googleProfile.Photo = profile.Photos[0].Url
	}

	redirectUrl := fmt.Sprintf("%s/auth/google/callback?name=%s&email=%s&phone=%s&birthday=%s&photo=%s",
		config.MonolithUrl+"/api/v1",
		googleProfile.Name,
		googleProfile.Email,
		googleProfile.Phone,
		googleProfile.Birthday,
		googleProfile.Photo,
	)

	ctx.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}

func getProfileFromGoogle(ctx *gin.Context, token *oauth2.Token) (*people.Person, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	service, err := people.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	profile, err := service.People.Get("people/me").PersonFields("names,emailAddresses,phoneNumbers,birthdays,photos").Do()
	if err != nil {
		return nil, err
	}
	return profile, nil
}

var facebookAdapter = auth.NewFacebookOAuth()

func FacebookLogin(ctx *gin.Context) {
	ctx.Redirect(http.StatusTemporaryRedirect, facebookAdapter.GenerateLoginURL())
}

func FacebookCallback(ctx *gin.Context) {
	token, err := facebookAdapter.LoginCallback(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", "Error getting token from Facebook", nil))
		return
	}

	/*
		TODO : Do something with the token, such as getting the user's email address
			using the Facebook API and redirect to client register page
	*/
	fmt.Println("Facebook token: %+v", token)
}

var microsoftAdapter = auth.NewMicrosoftOAuth()

func MicrosoftLogin(ctx *gin.Context) {
	ctx.Redirect(http.StatusTemporaryRedirect, microsoftAdapter.GenerateLoginURL())
}

func MicrosoftCallback(ctx *gin.Context) {
	token, err := microsoftAdapter.LoginCallback(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", "Error getting token from Microsoft", nil))
		return
	}

	// Use the token to access the user's profile.
	/*
		TODO : Do something with the token, such as getting the user's email address
			using the Microsoft API and redirect to client register page
	*/
	fmt.Println("Microsoft token: %+v", token)
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
		config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}

	refreshToken, refreshPayload, err := tokenMaker.CreateToken(
		int32(uId),
		config.RefreshTokenDuration,
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
		fmt.Println("customUser", customUser)
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
		otpCode, _ := utils.GenerateOTP(8)
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
		receiverList := []utils.EmailReceiver{
			{
				Name:    input["receiver"].(string),
				Address: input["receiver"].(string),
			},
		}

		// Send email verification
		go func() {
			resetToken := utils.Encode(fmt.Sprintf("email#%v#%v", otpCode, input["receiver"]))
			emailData := emailForgotPasswordParams{
				URL:     input["redirect_url"].(string) + resetToken,
				Name:    input["receiver"].(string),
				Subject: "Your password reset token (valid for 10min)",
			}
			fmt.Println("emailData", emailData)
			utils.EmailSender("forgot_password.html", emailData, receiverList)
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
	decodePasswordToken, err := utils.Decode(ctx.Param("token"))
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
	fmt.Println(hashedPassword)

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "reset password successfully", nil))
}

type privyRequest struct {
	ReferenceNumber string `json:"reference_number"`
	ChannelId       string `json:"channel_id"`
	Info            string `json:"info"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Nik             string `json:"nik"`
	Name            string `json:"name"`
	Dob             string `json:"dob"`
	Selfie          string `json:"selfie"`
	Identity        string `json:"identity"`
}

type privyResponse struct {
	ReferenceNumber string        `json:"reference_number"`
	RegisterToken   string        `json:"register_token"`
	Status          string        `json:"status"`
	ChannelId       string        `json:"channel_id"`
	PrivyId         string        `json:"privy_id"`
	Email           string        `json:"email"`
	Phone           string        `json:"phone"`
	Identity        privyIdentity `json:"identity"`
}

type privyIdentity struct {
	Nik          string `json:"nik"`
	Name         string `json:"name"`
	TanggalLahir string `json:"tanggal_lahir"`
}

func PrivyRegister(ctx *gin.Context) {
	var req privyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	jsonBody, _ := json.Marshal(req)
	var objBd map[string]interface{}
	if err := json.Unmarshal(jsonBody, &objBd); err != nil {
		panic(err)
	}

	privyAccessToken := privyAdapter.GetAccessToken()
	privyRegsCredentials := privyAdapter.GenerateCredentials(objBd)
	privyRegs := privyAdapter.RegisterUser(objBd, privyAccessToken, privyRegsCredentials["timestamp"], privyRegsCredentials["signature"], privyRegsCredentials["reference_number"])

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "success create privy user", privyRegs))
}

var privyAdapter = auth.NewPrivyOAuth()

func PrivyLogin(ctx *gin.Context) {
	ctx.Redirect(http.StatusTemporaryRedirect, privyAdapter.GenerateLoginURL())
}

func PrivyCallback(ctx *gin.Context) {
	config, err := utils.LoadConfig(".")
	if err != nil {
		fmt.Errorf("cannot load config: %w", err)
		return
	}

	token, err := privyAdapter.LoginCallback(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", "Error getting token from Privy", nil))
		return
	}

	profile, err := getProfileFromPrivy(config, token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", "Error getting data user from Privy", nil))
		return
	}

	profileJson, _ := json.Marshal(profile)
	redirectUrl := fmt.Sprintf("%s/auth/privy/callback?token=%v", config.MonolithUrl+"/api/v1", utils.Encode(string(profileJson)))

	ctx.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}

func getProfileFromPrivy(config utils.Config, token *oauth2.Token) (map[string]any, error) {
	// create HTTP request with POST method and request body
	req, err := http.NewRequest("GET", config.PrivyAuthGetUserExchangeUrl, nil)
	if err != nil {
		return nil, err
	}

	// set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	// send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read HTTP response body
	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	// print HTTP response status code and body
	return response, nil
}
