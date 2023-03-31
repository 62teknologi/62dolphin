package utils

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB1 *gorm.DB

var DB2 *gorm.DB

var DB *gorm.DB

func ConnectDatabase(cfg Config) {

	db1, err := gorm.Open(postgres.Open(cfg.DBSource1), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}

	DB1 = db1

	if cfg.DBSource2 != "" {
		db2, err := gorm.Open(postgres.Open(cfg.DBSource2), &gorm.Config{})
		if err != nil {
			panic("Failed to connect to database!")
		}
		DB2 = db2
	}

	DB = db1
}
