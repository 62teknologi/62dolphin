package main

import (
	"fmt"
	"github.com/62teknologi/62dolphin/62golib/utils"
	"github.com/62teknologi/62dolphin/app/config"
	"github.com/62teknologi/62dolphin/app/http/controllers"
	"github.com/62teknologi/62dolphin/app/http/middlewares"
	"github.com/62teknologi/62dolphin/app/tokens"

	"github.com/gin-gonic/gin"
)

func main() {

	configs, err := config.LoadConfig(".", &config.Data)
	if err != nil {
		fmt.Printf("cannot load config: %w", err)
		return
	}

	// todo : replace last variable with spread notation "..."
	utils.ConnectDatabase(configs.DBDriver, configs.DBSource1, configs.DBSource2)

	tokenMaker, err := tokens.NewJWTMaker(configs.TokenSymmetricKey)
	if err != nil {
		fmt.Printf("cannot create token maker: %w", err)
		return
	}

	r := gin.Default()

	pub := r.Use(middlewares.DbSelectorMiddleware())
	{
		pub.GET("/health", controllers.CheckAppHealth)

	}

	//todo use middleware db selector
	apiV1 := r.Group("/api/v1").Use(middlewares.DbSelectorMiddleware())
	{
		apiV1.POST("/auth/sign-in", controllers.Callback)
		apiV1.POST("/auth/sign-up", controllers.CreateUser)

		// adapter : local, facebook, microsoft, google
		apiV1.GET("/auth/:adapter", controllers.Login)
		apiV1.GET("/auth/:adapter/callback", controllers.Callback)
		apiV1.POST("/auth/:adapter/callback", controllers.Callback)

		apiV1.POST("/otps/create", controllers.CreateOTP)

		apiV1.POST("/tokens/create", controllers.CreateAccessToken)
		apiV1.POST("/tokens/verify", controllers.VerifyAccessToken)
		apiV1.POST("/tokens/refresh", controllers.RenewAccessToken)

		apiV1.POST("/passwords/create", controllers.CreateHashPassword)
		apiV1.POST("/passwords/check", controllers.CheckPassword)
		apiV1.POST("/passwords/forgot", controllers.ForgotPassword)
		apiV1.PATCH("/passwords/reset/:token", controllers.ResetPassword)

		apiV1.GET("/users", controllers.FindUsers)
		apiV1.POST("/users", controllers.CreateUser)
		apiV1.GET("/users/:id", controllers.FindUser)
		apiV1.POST("/users/verify", controllers.VerifyUser)
	}

	authorizedV1 := r.Group("/api/v1").Use(middlewares.AuthMiddleware(tokenMaker))
	{
		authorizedV1.POST("/tokens/block", controllers.BlockRefreshToken)
		authorizedV1.POST("/tokens/block-all", controllers.BlockAllRefreshToken)
		authorizedV1.PUT("/users/:id", controllers.UpdateUser)
		authorizedV1.DELETE("/users/:id", controllers.DeleteUser)
	}

	err = r.Run(configs.HTTPServerAddress)
	if err != nil {
		fmt.Printf("cannot run server: %w", err)
		return
	}
}
