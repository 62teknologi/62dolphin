package routes

import (
	"dolphin/app/http/controllers"
	"github.com/gin-gonic/gin"
)

func PasswordRoutesV1(route *gin.RouterGroup) {

	passwords := route.Group("/passwords")
	{
		passwords.POST("/forgot-password", controllers.ForgotPassword)
		passwords.PATCH("/reset-password/:token", controllers.ResetPassword)
	}
}
