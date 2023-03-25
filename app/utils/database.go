package utils

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase(dbSource string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dbSource), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}

	//err = db.Debug().AutoMigrate(User{}, Token{})
	//if err != nil {
	//	panic("migration error")
	//}

	DB = db
	return db
}
