package main

import (
	"dolphin/app/http/controllers"
	"dolphin/app/models"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	models.ConnectDatabase()

	r.GET("/sign-in", controllers.SignIn)
	r.GET("/users", controllers.FindUsers)
	r.POST("/users", controllers.CreateUser)
	r.GET("/users/:id", controllers.FindUser)
	r.PUT("/users/:id", controllers.UpdateUser)
	r.DELETE("/users/:id", controllers.DeleteUser) // new

	err := r.Run()
	if err != nil {
		return
	}
}
