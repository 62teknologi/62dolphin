package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"dolphin/app/utils"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"golang.org/x/oauth2"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type PrivyOAuth struct {
	oauthConfig *oauth2.Config
}

func NewPrivyOAuth() *PrivyOAuth {
	return &PrivyOAuth{
		oauthConfig: &oauth2.Config{
			ClientID:     config.PrivyAuthClientId,
			ClientSecret: config.PrivyAuthClientSecret,
			RedirectURL:  config.PrivyAuthRedirectUrl,
			Scopes:       []string{"read", "write"},
			Endpoint: oauth2.Endpoint{
				AuthURL: config.PrivyAuthUrl,
			}},
	}
}

func (p *PrivyOAuth) Exchange(_ *gin.Context, code string) (*oauth2.Token, error) {
	// create request body as a JSON-encoded byte slice
	requestBody, err := json.Marshal(map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     config.PrivyAuthClientId,
		"client_secret": config.PrivyAuthClientSecret,
		"redirect_uri":  config.PrivyAuthRedirectUrl,
		"code":          code,
	})
	if err != nil {
		return nil, err
	}

	// create HTTP request with POST method and request body
	req, err := http.NewRequest("POST", config.PrivyAuthTokenExchangeUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	// set request headers
	req.Header.Set("Content-Type", "application/json")

	// send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read HTTP response body
	var response *oauth2.Token
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	// print HTTP response status code and body
	return response, nil
}

func (p *PrivyOAuth) GenerateLoginURL() string {
	url := p.oauthConfig.AuthCodeURL("")
	return url
}

func (p *PrivyOAuth) LoginCallback(ctx *gin.Context) (*oauth2.Token, error) {
	code := ctx.Query("code")
	token, err := p.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (p *PrivyOAuth) GetAccessToken() string {
	authClient := resty.New()
	authResp, err := authClient.R().
		SetBody(map[string]string{
			"client_id":     config.PrivyApiKey,
			"client_secret": config.PrivySecretKey,
			"grant_type":    "client_credentials",
		}).
		Post(config.PrivyUrl + "/oauth2/api/v1/token")

	if err != nil {
		panic(err)
	}

	var authRespBody map[string]any
	json.Unmarshal(authResp.Body(), &authRespBody)
	accessToken := authRespBody["data"].(map[string]any)["access_token"].(string)

	return accessToken
}

func (p *PrivyOAuth) GenerateCredentials(body map[string]any, referenceNumber string) map[string]string {
	// Generate reference number and request number
	rand.Seed(time.Now().UnixNano())
	num := rand.Int63n(10000000000000000)

	if referenceNumber == "" {
		referenceNumber = fmt.Sprintf("%016d", num)
	}

	timestamp := time.Now().Format("Mon, 02 Jan 2006 15:04:05 GMT")

	// Set reference number and channel id from environment variables
	body["reference_number"] = referenceNumber
	body["channel_id"] = config.PrivyChannelId

	// Remove selfie and identity properties from the JSON object
	delete(body, "user_id")
	delete(body, "selfie")
	delete(body, "identity")

	method := "POST"
	jsonBd, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	jsonBdStr := strings.ReplaceAll(string(jsonBd), " ", "")
	jsonBdStr = strings.ReplaceAll(jsonBdStr, "\n", "")

	bodyMd5 := md5.Sum([]byte(jsonBdStr))
	bodyMd5Base64 := base64.StdEncoding.EncodeToString(bodyMd5[:])

	hmacSignature := fmt.Sprintf("%s:%s:%s:%s", timestamp, config.PrivyApiKey, method, bodyMd5Base64)
	hmacHash := hmac.New(sha256.New, []byte(config.PrivySecretKey))
	hmacHash.Write([]byte(hmacSignature))
	hmacBase64 := base64.StdEncoding.EncodeToString(hmacHash.Sum(nil))

	keyString := fmt.Sprintf("%s:%s", config.PrivyApiKey, hmacBase64)
	signature := base64.StdEncoding.EncodeToString([]byte(keyString))
	return map[string]string{
		"timestamp":        timestamp,
		"signature":        signature,
		"reference_number": referenceNumber,
	}
}

func (p *PrivyOAuth) RegisterUser(body map[string]any, accessToken string, timestamp string, signature string, referenceNumber string) map[string]any {
	client := resty.New()
	regResp, err := client.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", accessToken)).
		SetHeader("Timestamp", timestamp).
		SetHeader("Signature", signature).
		SetHeader("Request-ID", referenceNumber).
		SetBody(body).
		Post(config.PrivyUrl + "/web/api/v2/register")

	if err != nil {
		panic(err)
	}

	var regRespBody map[string]any
	json.Unmarshal(regResp.Body(), &regRespBody)
	regsInfo := regRespBody["data"].(map[string]any)

	_ = utils.DB.Table("privy_register_logs").Create(map[string]any{
		"user_email":       body["email"],
		"reference_number": regsInfo["reference_number"],
		"register_token":   regsInfo["register_token"],
		"status":           regsInfo["status"],
		"registration_url": regsInfo["registration_url"],
	})

	return regsInfo
}

func (p *PrivyOAuth) RegisterStatus(body map[string]any, accessToken string, timestamp string, signature string, referenceNumber string) map[string]any {
	client := resty.New()
	regResp, err := client.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", accessToken)).
		SetHeader("Timestamp", timestamp).
		SetHeader("Signature", signature).
		SetHeader("Request-ID", referenceNumber).
		SetBody(body).
		Post(config.PrivyUrl + "/web/api/v2/register/status")

	if err != nil {
		panic(err)
	}

	var regRespBody map[string]any
	json.Unmarshal(regResp.Body(), &regRespBody)
	regsInfo := regRespBody["data"].(map[string]any)

	return regsInfo
}
