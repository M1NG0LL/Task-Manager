package main

import (
	account_package "project/Account"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	router := gin.Default()

	var err error
	db, err = gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate 
	if err := db.AutoMigrate(&account_package.Account{}); err != nil {
		panic("failed to migrate database")
	}

	// Account Part
	account_package.Init(db)

	router.POST("/login", account_package.Login)
	router.POST("/accounts", account_package.CreateAccount)
	router.PUT("/activation/:id", account_package.ActivateAccountByID)

	protected := router.Group("/")
	protected.Use(account_package.AuthMiddleware())

	protected.GET("/accounts", account_package.GetMyAccount)
	protected.PUT("/accounts/:id", account_package.UpdateAccountByID)
	protected.DELETE("/accounts/:id", account_package.DeleteAccountbyid)

	router.Run(":8081")
}