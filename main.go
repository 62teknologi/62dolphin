package main

import (
	"dolphin/app/http/controllers"
	"dolphin/app/http/middlewares"
	"dolphin/app/tokens"
	"dolphin/app/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		fmt.Printf("cannot load config: %w", err)
		return
	}

	db := utils.ConnectDatabase(config.DBSource)

	tokenMaker, err := tokens.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		fmt.Printf("cannot create token maker: %w", err)
		return
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		dbConn, _ := db.DB()
		parsedDsn, _ := url.Parse(config.DBSource)
		host := parsedDsn.Host
		dbName := parsedDsn.Path

		if host == "" {
			// Parse DSN server format
			pairs := strings.Split(dbName, " ")
			data := make(map[string]string)
			for _, pair := range pairs {
				parts := strings.Split(pair, "=")
				if len(parts) == 2 {
					data[parts[0]] = parts[1]
				}
			}
			host = data["host"] + ":" + data["port"]
			dbName = data["dbname"]
		}

		if err := dbConn.Ping(); err != nil {
			c.JSON(http.StatusOK, utils.ResponseData("success", "Server running well", map[string]any{
				"server_status":   "ok",
				"database_status": "error",
				"database_name":   dbName,
				"database_host":   host,
			}))
			return
		}

		c.JSON(http.StatusOK, utils.ResponseData("success", "Server running well", map[string]any{
			"server_status":   "ok",
			"database_status": "ok",
			"database_name":   dbName,
			"database_host":   host,
		}))
	})

	apiV1 := r.Group("/api/v1")
	{
		/*
			Auth
		*/
		apiV1.POST("/auth/sign-in", controllers.SignIn)
		apiV1.POST("/auth/sign-up", controllers.CreateUser)

		apiV1.GET("/auth/google", controllers.GoogleLogin)
		// TODO need to change to apiV1
		r.GET("/auth/callback/google", controllers.GoogleCallback)

		apiV1.GET("/auth/facebook", controllers.FacebookLogin)
		apiV1.GET("/auth/callback/facebook", controllers.FacebookCallback)

		apiV1.GET("/auth/microsoft", controllers.MicrosoftLogin)
		apiV1.GET("/auth/callback/microsoft", controllers.MicrosoftCallback)

		apiV1.POST("/auth/privy/register", controllers.PrivyRegister)
		apiV1.GET("/auth/privy/register/otp", controllers.PrivyOtp)
		apiV1.POST("/auth/privy/register/status", controllers.PrivyRegisterStatus)
		apiV1.GET("/auth/privy", controllers.PrivyLogin)
		apiV1.GET("/auth/privy/callback", controllers.PrivyCallback)

		/*
			Tokens
		*/
		apiV1.POST("/tokens/create", controllers.CreateAccessToken)
		apiV1.POST("/tokens/verify", controllers.VerifyAccessToken)
		apiV1.POST("/tokens/refresh", controllers.RenewAccessToken)

		/*
			Passwords
		*/
		apiV1.POST("/passwords/forgot-password", controllers.ForgotPassword)
		apiV1.PATCH("/passwords/reset-password/:token", controllers.ResetPassword)

		/*
			Users
		*/
		apiV1.GET("/users", controllers.FindUsers)
		apiV1.POST("/users", controllers.CreateUser)
		apiV1.GET("/users/:id", controllers.FindUser)
		apiV1.POST("/users/verify", controllers.VerifyUser)
	}

	authorizedV1 := r.Group("/api/v1").Use(middlewares.AuthMiddleware(tokenMaker))
	{
		/*
			Tokens
		*/
		authorizedV1.POST("/tokens/block-token", controllers.BlockRefreshToken)

		/*
			Users
		*/
		authorizedV1.PUT("/users/:id", controllers.UpdateUser)
		authorizedV1.DELETE("/users/:id", controllers.DeleteUser)
	}

	err = r.Run(config.HTTPServerAddress)
	if err != nil {
		fmt.Printf("cannot run server: %w", err)
		return
	}
}
