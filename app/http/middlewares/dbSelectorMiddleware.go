package middlewares

import (
	"dolphin/app/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates a gin middleware for authorization
func DbSelectorMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		if ctx.Query("db") == "2" {
			utils.DB = utils.DB2
		} else {
			utils.DB = utils.DB1
		}
		//DB = ctx.query(db) ??  DB1;

		// ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
