package main

import (
	"dolphin/app/http/controllers"
	"dolphin/app/http/middlewares"
	"dolphin/app/models"
	"dolphin/app/tokens"
	"dolphin/app/utils"
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		fmt.Printf("cannot load config: %w", err)
		return
	}

	models.ConnectDatabase(config.DBSource)

	r := gin.Default()
	r.POST("/auth/sign-in", controllers.SignIn)
	r.GET("/users", controllers.FindUsers)
	r.POST("/users", controllers.CreateUser)
	r.GET("/users/:id", controllers.FindUser)

	tokenMaker, err := tokens.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		fmt.Printf("cannot create token maker: %w", err)
		return
	}
	authorized := r.Group("/").Use(middlewares.AuthMiddleware(tokenMaker))
	authorized.PUT("/users/:id", controllers.UpdateUser)
	authorized.DELETE("/users/:id", controllers.DeleteUser)

	err = r.Run()
	if err != nil {
		fmt.Printf("cannot run server: %w", err)
		return
	}
}
