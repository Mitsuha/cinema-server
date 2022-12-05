package model

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DBConfig struct {
	Username  string `yaml:"Username"`
	Password  string `yaml:"Password"`
	Localhost string `yaml:"Localhost"`
	Port      int    `yaml:"Port"`
	Database  string `yaml:"Database"`
}

var DB *gorm.DB

func Boot(config *DBConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.Username, config.Password, config.Localhost, config.Port, config.Database)

	if db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{}); err != nil {
		return err
	} else {
		DB = db
	}
	return nil
}
