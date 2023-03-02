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
