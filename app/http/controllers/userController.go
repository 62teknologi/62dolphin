package controllers

import (
	"dolphin/app/utils"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dbssensei/ordentmarketplace/util"
	"github.com/go-sql-driver/mysql"

	"github.com/gin-gonic/gin"
)

func FindUser(ctx *gin.Context) {
	// query to find user
	var user map[string]interface{}
	err := utils.DB.Table("users").Where("is_active", true).Where("id = ?", ctx.Param("id")).Take(&user).Error
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}
	if user["id"] == nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "user not found", nil))
		return
	}

	// Setup output to client
	customResponse, err := utils.JsonFileParser("transformers/response/user/get.json")
	customUser := customResponse["user"]

	utils.MapValuesShifter(customResponse, user)
	if customUser != nil {
		utils.MapValuesShifter(customUser.(map[string]any), user)
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "find user successfully", customResponse))
}

func FindUsers(ctx *gin.Context) {
	var users []map[string]interface{}
	err := utils.DB.Table("users").Find(&users).Error
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}

	var customResponses []map[string]any
	for _, user := range users {
		// Setup output to client
		customResponse, _ := utils.JsonFileParser("transformers/response/user/get.json")
		customUser := customResponse["user"]

		utils.MapValuesShifter(customResponse, user)
		if customUser != nil {
			utils.MapValuesShifter(customUser.(map[string]any), user)
		}
		customResponses = append(customResponses, customResponse)
	}

	// return collection of transformed user
	ctx.JSON(http.StatusOK, utils.ResponseData("success", "find all users successfully", customResponses))
}

type otpVerificationParams struct {
	OtpReceiver string
	OtpCode     string
}

func CreateUser(ctx *gin.Context) {
	// Parse and cleaning input
	input, err := utils.JsonFileParser("transformers/request/user/create.json")
	var userInput map[string]any
	if err = ctx.BindJSON(&userInput); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}
	utils.MapValuesShifter(input, userInput)
	utils.MapNullValuesRemover(input)

	// Move all otp keys on input to difference map
	otpOptions := make(map[string]any)
	for k, v := range input {
		if strings.Contains(k, "otp") {
			otpOptions[k] = v
			delete(input, k)
		}
	}

	// Hashing Password
	hashedPassword, err := util.HashPassword(input["password"].(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", fmt.Sprintf("%v", err.Error()), nil))
		return
	}

	// Set default fields
	input["password"] = hashedPassword
	//input["created_at"] = time.Now()
	//input["updated_at"] = time.Now()
	if otpOptions["otp"] == true {
		input["is_active"] = false
	}

	// Create and handle query error
	createUserQuery := utils.DB.Table("users").Create(input)
	if createUserQuery.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(createUserQuery.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", createUserQuery.Error.Error(), nil))
			return
		}
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", createUserQuery.Error.Error(), nil))
		return
	}

	// Generate and create OTP if otp option is active
	if otpOptions["otp"] == true {
		fmt.Println("otpOptions", otpOptions)
		otpCode, _ := utils.GenerateOTP(8)
		otpParams := map[string]any{
			"type":       otpOptions["otp_method"],
			"code":       otpCode,
			"receiver":   otpOptions["otp_receiver"],
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
				Name:    "Dimas",
				Address: input["email"].(string),
			},
		}

		// Send email verification
		fmt.Printf("%+v\n", otpOptions)
		go func() {
			utils.EmailSender("verify_user.html", otpVerificationParams{OtpReceiver: otpOptions["otp_receiver"].(string), OtpCode: otpCode}, receiverList)
		}()
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "create user successfully", nil))
}

func VerifyUser(ctx *gin.Context) {
	// Parse and cleaning input
	input, err := utils.JsonFileParser("transformers/request/user/verify.json")
	var userInput map[string]any
	if err = ctx.BindJSON(&userInput); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}
	utils.MapValuesShifter(input, userInput)
	utils.MapNullValuesRemover(input)

	// Check if user exist in db
	var otp map[string]any
	utils.DB.Table("otps").Where("type = ?", input["method"]).Where("receiver = ?", input["receiver"]).Where("code = ?", input["code"]).Take(&otp)

	if otp["id"] == nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "invalid otp", nil))
		return
	}

	//if otp["expires_at"].(time.Time).Unix() < time.Now().Unix() {
	//	ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "otp expired", nil))
	//	return
	//}

	// Update and handle query error
	params := map[string]any{
		"is_active": true,
		//"updated_at": time.Now(),
	}
	fmt.Println("input", input)

	var user map[string]any
	utils.DB.Table("users").Where(fmt.Sprintf("%v = ?", input["method"]), input["receiver"]).Take(&user)
	fmt.Println("user", user)
	updateResultQuery := utils.DB.Table("users").Where(fmt.Sprintf("%v = ?", input["method"]), input["receiver"]).Updates(&params)
	fmt.Printf("%+v", updateResultQuery)
	if updateResultQuery.Error != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", updateResultQuery.Error.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "verify user successfully", nil))
}

func UpdateUser(ctx *gin.Context) {
	// Parse and cleaning input
	input, err := utils.JsonFileParser("transformers/request/user/update.json")
	var userInput map[string]any
	if err = ctx.BindJSON(&userInput); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}
	utils.MapValuesShifter(input, userInput)
	utils.MapNullValuesRemover(input)

	// Hashing Password (if exist)
	if input["password"] != nil {
		hashedPassword, err := util.HashPassword(input["password"].(string))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", fmt.Sprintf("%v", err.Error()), nil))
			return
		}
		input["password"] = hashedPassword
	}

	// TODO verify updated data with otp

	// Set default fields
	//input["updated_at"] = time.Now()

	// Update and handle query error
	updateResultQuery := utils.DB.Table("users").Where("id = ?", ctx.Param("id")).Updates(input)
	if updateResultQuery.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(updateResultQuery.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", updateResultQuery.Error.Error(), nil))
			return
		}
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", updateResultQuery.Error.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "update user successfully", nil))
}

func DeleteUser(c *gin.Context) {}
