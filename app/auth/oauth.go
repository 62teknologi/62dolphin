package auth

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

// OAuth interface for managing 3pl auth from oauth
type OAuth interface {
	// GenerateLoginURL generate redirect for user auth
	GenerateLoginURL() string

	// LoginCallback handle user authentication after login success
	LoginCallback(ctx *gin.Context) (*oauth2.Token, error)
}

type OAuthProfile struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Birthday string `json:"birthday"`
	Photo    string `json:"photo"`
}
