package controllers

import (
	"github.com/62teknologi/62dolphin/62golib/utils"
	"github.com/dbssensei/ordentmarketplace/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

type createHashPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

func CreateHashPassword(ctx *gin.Context) { // Setup request body
	var req createHashPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	// Hashing Password
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "error while hashing password", nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "success create hash password", map[string]any{
		"hashed_password": hashedPassword,
	}))
}

type checkPasswordRequest struct {
	Password       string `json:"password" binding:"required"`
	HashedPassword string `json:"hashed_password" binding:"required"`
}

func CheckPassword(ctx *gin.Context) { // Setup request body
	var req checkPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	// Hashing Password
	err := util.CheckPassword(req.Password, req.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", "password does not match", nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "password match", nil))
}
