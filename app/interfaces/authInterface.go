package interfaces

import (
	"github.com/gin-gonic/gin"
)

type OAuthData struct {
	Email  string `json:"email"`
	UserId string `json:"user_id"`
}

type AuthInterface interface {
	Init() AuthInterface
	GenerateLoginURL() string
	Callback(ctx *gin.Context) error
	Verify(ctx *gin.Context) (map[string]any, error)
}
