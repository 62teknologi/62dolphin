package interfaces

import (
	"github.com/gin-gonic/gin"
)

type AuthInterface interface {
	Init() AuthInterface
	GenerateLoginURL() string
	Callback(ctx *gin.Context) error
}
