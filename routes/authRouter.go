package routes

import (
	"dolphin/app/http/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRoutesV1(route *gin.RouterGroup) {
	auth := route.Group("/auth")
	{
		auth.POST("/sign-in", controllers.SignIn)
		auth.POST("/sign-up", controllers.CreateUser)

		auth.GET("/google", controllers.GoogleLogin)
		auth.GET("/callback/google", controllers.GoogleCallback)

		auth.GET("/facebook", controllers.FacebookLogin)
		auth.GET("/callback/facebook", controllers.FacebookCallback)

		auth.GET("/microsoft", controllers.MicrosoftLogin)
		auth.GET("/callback/microsoft", controllers.MicrosoftCallback)
	}
}
