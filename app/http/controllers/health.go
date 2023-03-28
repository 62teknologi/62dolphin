package controllers

import (
	"dolphin/app/utils"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func CheckAppHealth(c *gin.Context) {
	config, _ := utils.LoadConfig(".")

	dbConn, _ := utils.DB.DB()
	parsedDsn, _ := url.Parse(config.DBSource1)

	if c.Query("db") == "2" {
		parsedDsn, _ = url.Parse(config.DBSource2)
	}

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
}
