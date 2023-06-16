package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/62teknologi/62dolphin/62golib/utils"
	"github.com/62teknologi/62dolphin/app/config"
	dutils "github.com/62teknologi/62dolphin/app/utils"

	"github.com/dbssensei/ordentmarketplace/util"

	"github.com/gin-gonic/gin"
)

func FindUser(ctx *gin.Context) {
	value := map[string]any{}
	columns := []string{"users.*"}
	order := "id desc"
	transformer, _ := utils.JsonFileParser(config.Data.SettingPath + "/transformers/response/users/find.json")
	query := utils.DB.Table("users")

	utils.SetBelongsTo(query, transformer, &columns)
	delete(transformer, "filterable")

	if err := query.Select(columns).Order(order).Where("users."+"id = ?", ctx.Param("id")).Take(&value).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "user not found", nil))
		return
	}

	utils.MapValuesShifter(transformer, value)
	utils.AttachBelongsTo(transformer, value)

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "find user success", transformer))
}

func FindUsers(ctx *gin.Context) {
	values := []map[string]any{}
	columns := []string{"users.*"}
	transformer, _ := utils.JsonFileParser(config.Data.SettingPath + "/transformers/response/users/find.json")
	query := utils.DB.Table("users")
	filter := utils.SetFilterByQuery(query, transformer, ctx)
	search := utils.SetGlobalSearch(query, transformer, ctx)

	utils.SetOrderByQuery(query, ctx)
	utils.SetBelongsTo(query, transformer, &columns)

	delete(transformer, "filterable")
	delete(transformer, "searchable")

	pagination := utils.SetPagination(query, ctx)

	if err := query.Select(columns).Find(&values).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "users not found", nil))
		return
	}

	customResponses := utils.MultiMapValuesShifter(transformer, values)
	summary := utils.GetSummary(transformer, values)

	ctx.JSON(http.StatusOK, utils.ResponseDataPaginate("success", "find users success", customResponses, pagination, filter, search, summary))
}

type otpVerificationParams struct {
	OtpReceiver string
	OtpCode     string
}

func CreateUser(ctx *gin.Context) {
	transformer, _ := utils.JsonFileParser(config.Data.SettingPath + "/transformers/request/users/create.json")
	input := utils.ParseForm(ctx)

	if validation, err := utils.Validate(input, transformer); err {
		ctx.JSON(http.StatusOK, utils.ResponseData("failed", "validation", validation.Errors))
		return
	}

	utils.MapValuesShifter(transformer, input)
	utils.MapNullValuesRemover(transformer)

	// Hashing Password
	hashedPassword, err := util.HashPassword(transformer["password"].(string))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", fmt.Sprintf("%v", err.Error()), nil))
		return
	}

	// Set default fields
	transformer["password"] = hashedPassword

	if input["otp"] == true && transformer["is_active"] == "" {
		transformer["is_active"] = false
	}

	// Create and handle query error
	if err := utils.DB.Table("users").Create(&transformer).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	// Generate and create OTP if otp option is active
	if input["otp"] == true {
		otpCode, _ := dutils.GenerateOTP(8)

		otpParams := map[string]any{
			"type":       input["otp_method"],
			"code":       otpCode,
			"receiver":   input["otp_receiver"],
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
				Name:    "Dimas",
				Address: transformer["email"].(string),
			},
		}

		// Send email verification
		go func() {
			dutils.EmailSender("verify_user.html", otpVerificationParams{OtpReceiver: input["otp_receiver"].(string), OtpCode: otpCode}, receiverList)
		}()
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "create user success", transformer))
}

func VerifyUser(ctx *gin.Context) {
	// Parse and cleaning input
	transformer, _ := utils.JsonFileParser(config.Data.SettingPath + "/transformers/request/users/verify.json")
	input := utils.ParseForm(ctx)

	if validation, err := utils.Validate(input, transformer); err {
		ctx.JSON(http.StatusOK, utils.ResponseData("failed", "validation", validation.Errors))
		return
	}

	utils.MapValuesShifter(transformer, input)
	utils.MapNullValuesRemover(transformer)

	// Check if user exist in db
	var otp map[string]any
	utils.DB.Table("otps").Where("type = ?", transformer["method"]).Where("receiver = ?", transformer["receiver"]).Where("code = ?", transformer["code"]).Take(&otp)

	if otp["id"] == nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "invalid otp", nil))
		return
	} else if otp["expires_at"].(time.Time).Unix() < time.Now().Unix() {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "otp expired", nil))
		return
	}

	params := map[string]any{
		"is_active": true,
	}

	user := map[string]any{}
	utils.DB.Table("users").Where(fmt.Sprintf("%v = ?", transformer["method"]), transformer["receiver"]).Take(&user)
	err := utils.DB.Table("users").Where(fmt.Sprintf("%v = ?", transformer["method"]), transformer["receiver"]).Updates(&params).Error

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "verify user successfully", nil))
}

// TODO verify updated data with otp
func UpdateUser(ctx *gin.Context) {
	transformer, _ := utils.JsonFileParser(config.Data.SettingPath + "/transformers/request/users/update.json")
	input := utils.ParseForm(ctx)

	if validation, err := utils.Validate(input, transformer); err {
		ctx.JSON(http.StatusOK, utils.ResponseData("failed", "validation", validation.Errors))
		return
	}

	utils.MapValuesShifter(transformer, input)
	utils.MapNullValuesRemover(transformer)

	// Hashing Password (if exist)
	if input["password"] != nil {
		hashedPassword, err := util.HashPassword(input["password"].(string))

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", fmt.Sprintf("%v", err.Error()), nil))
			return
		}

		transformer["password"] = hashedPassword
	}

	if err := utils.DB.Table("users").Where("id = ?", ctx.Param("id")).Updates(transformer).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "update user success", transformer))
}

func DeleteUser(c *gin.Context) {}
