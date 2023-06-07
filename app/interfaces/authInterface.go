package interfaces

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type AuthInterface interface {
	Init() AuthInterface
	GenerateLoginURL() string
	Callback(ctx *gin.Context) (*oauth2.Token, error)
}

type AdapterProfile struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Birthday string `json:"birthday"`
	Photo    string `json:"photo"`
}
