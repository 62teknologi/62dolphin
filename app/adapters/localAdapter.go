package adapters

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/62teknologi/62dolphin/62golib/utils"
	"github.com/62teknologi/62dolphin/app/config"
	"github.com/62teknologi/62dolphin/app/interfaces"
	"github.com/62teknologi/62dolphin/app/tokens"
	dutils "github.com/62teknologi/62dolphin/app/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

type LocalAdapter struct {
	config *oauth2.Config
}

func (adp *LocalAdapter) Init() interfaces.AuthInterface {
	adp.config = &oauth2.Config{
		ClientID:     config.Data.FacebookAuthClientId,
		ClientSecret: config.Data.FacebookAuthClientSecret,
		RedirectURL:  config.Data.FacebookAuthRedirectUrl,
		Scopes:       []string{"email"},
		Endpoint:     facebook.Endpoint,
	}

	return adp
}

func (adp *LocalAdapter) GenerateLoginURL() string {
	return "/api/v1/auth/local/callback"
}

func (adp *LocalAdapter) Verify(ctx *gin.Context, email, userId string) (map[string]any, error) {
	return map[string]any{"status": "success"}, nil
}

func (adp *LocalAdapter) Callback(ctx *gin.Context) error {
	transformer, err := utils.JsonFileParser(config.Data.SettingPath + "/transformers/request/auth/login.json")
	input := utils.ParseForm(ctx)

	authField := transformer["auth_field"]
	delete(transformer, "auth_field")

	// customQuery := transformer["custom_query"]
	// delete(transformer, "custom_query")

	if validation, err := utils.Validate(input, transformer); err {
		utils.LogJson(validation.Errors)
		return errors.New("validation error")
	}

	utils.MapValuesShifter(transformer, input)
	utils.MapNullValuesRemover(transformer)

	tokenMaker, err := tokens.NewJWTMaker(config.Data.TokenSymmetricKey)

	if err != nil {
		fmt.Errorf("cannot create token maker: %w", err)
	}

	ctx.Set("transformer", transformer)
	ctx.Set("input", input)
	ctx.Set("auth_field", authField)

	user, err := adp.getProfile(ctx)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return err
	}

	uId, _ := strconv.ParseInt(fmt.Sprintf("%v", user["id"]), 10, 32)

	accessToken, accessPayload, err := tokenMaker.CreateToken(
		int32(uId),
		config.Data.AccessTokenDuration,
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
	}

	refreshToken, refreshPayload, err := tokenMaker.CreateToken(
		int32(uId),
		config.Data.RefreshTokenDuration,
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", err.Error(), nil))
	}

	// Store sessions data to DB
	params := map[string]any{
		"id":            refreshPayload.Id,
		"user_id":       int32(uId),
		"refresh_token": refreshToken,
		"platform_id":   int32(transformer["platform_id"].(float64)),
		"is_blocked":    false,
		"expires_at":    refreshPayload.ExpiredAt,
		"created_at":    time.Now(),
		"updated_at":    time.Now(),
	}

	if err := utils.DB.Table("tokens").Create(&params).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ResponseData("error", fmt.Sprintf("%v", err.Error()), nil))
	}

	defaultResponse := map[string]any{
		"session_id":               params["id"],
		"access_token":             accessToken,
		"access_token_expires_at":  accessPayload.ExpiredAt,
		"refresh_token":            refreshToken,
		"refresh_token_expires_at": refreshPayload.ExpiredAt,
		"platform_id":              params["platform_id"],
	}

	customResponse, err := utils.JsonFileParser(config.Data.SettingPath + "/transformers/response/auth/login.json")
	customUser := customResponse["user"]

	utils.MapValuesShifter(customResponse, defaultResponse)

	if customUser != nil {
		utils.MapValuesShifter(customUser.(map[string]any), user)
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "sign-in successfully", customResponse))

	return nil
}

func (adp *LocalAdapter) getProfile(ctx *gin.Context) (map[string]any, error) {
	transformer := ctx.MustGet("transformer").(map[string]any)
	input := ctx.MustGet("input").(map[string]any)
	authField := ctx.MustGet("auth_field").(string)

	var query *gorm.DB
	query = utils.DB.Table("users")

	if strings.Contains(authField, "|") {
		substrings := strings.Split(authField, "|")
		var conditions []string
		var values []interface{}

		for _, substring := range substrings {
			if value, ok := transformer[substring]; ok {
				conditions = append(conditions, fmt.Sprintf("%s = ?", substring))
				values = append(values, value)
			}
		}

		// membuat group condition WHERE (email = 'implementer@email.com' OR username = 'username')
		if len(conditions) > 0 {
			rawQuery := strings.Join(conditions, " OR ")
			query = query.Where(rawQuery, values...)

		}
	} else {
		query = query.Where(utils.DB.Where(fmt.Sprintf("%s = ?", authField), transformer[authField]))
	}

	// Menambahkan custom_query dengan berbagai field
	if customQuery, ok := transformer["custom_query"].(map[string]any); ok {
		for field, value := range customQuery {
			var convertedValue any
			switch v := value.(type) {
			case string:
				loweredValue := strings.ToLower(v)
				if loweredValue == "true" || loweredValue == "1" {
					convertedValue = true
				} else if loweredValue == "false" || loweredValue == "0" {
					convertedValue = false
				} else {
					convertedValue = v
				}
			case int:
				if v == 1 {
					convertedValue = true
				} else if v == 0 {
					convertedValue = false
				} else {
					convertedValue = v
				}
			case bool:
				convertedValue = v
			default:
				convertedValue = v
			}
			query = query.Where(fmt.Sprintf("%s = ?", field), convertedValue)
		}
	}

	// is user not deleted_at
	query = query.Where("deleted_at IS NULL")
	query.Take(&transformer)

	if transformer["id"] == nil {
		return transformer, fmt.Errorf("user in not registered on inactive", authField)
	}

	if err := dutils.CheckPassword(input["password"].(string), transformer["password"].(string)); err != nil {
		return transformer, fmt.Errorf("invalid %s or password", authField)
	}

	return transformer, nil
}
