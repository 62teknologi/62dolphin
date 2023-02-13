package models

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase(dbSource string) {
	db, err := gorm.Open(mysql.Open(dbSource), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}

	err = db.Debug().AutoMigrate(User{}, Token{})
	if err != nil {
		panic("migration error")
	}

	DB = db
}
