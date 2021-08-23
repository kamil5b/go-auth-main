package database

import (
	"github.com/kamil5b/go-auth-main/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

const DNS = "root:@/programakuntansi?parseTime=true"

func Connect() {
	connection, err := gorm.Open(mysql.Open(DNS), &gorm.Config{})

	if err != nil {
		panic("could not connect to the database")
	}

	DB = connection
	connection.AutoMigrate(&models.User{})
}
