package main

import (
	account_package "project/Account"
	task_package "project/Task"

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

	router.POST("/login", account_package.Login) // good
	router.POST("/accounts", account_package.CreateAccount) // good
	router.PUT("/activation/:id", account_package.ActivateAccountByID) // good

	protected := router.Group("/")
	protected.Use(account_package.AuthMiddleware())

	protected.GET("/accounts", account_package.GetMyAccount) // good
	protected.PUT("/accounts/:id", account_package.UpdateAccountByID) // good
	protected.DELETE("/accounts/:id", account_package.DeleteAccountbyid) // good


	// Task Manager part

	if err := db.AutoMigrate(&task_package.Tasks{}, &task_package.To_DO_Tasks{}); err != nil {
		panic("failed to migrate database")
	}

	task_package.InitializeDB(db)
	protected.POST("/accounts/Tasks", task_package.CreateTask) // good
	protected.POST("/accounts/Tasks/:id", task_package.CreateTaskbyID) // good
	protected.GET("/accounts/Tasks", task_package.GetMyTasks) // GETTASKS have problem in printing 
	protected.GET("/accounts/Tasks/:id", task_package.GetMyTasksbyID) // good
	protected.PUT("/accounts/Tasks", task_package.UpdateMyTask) //good
	protected.PUT("/accounts/Tasks/:id", task_package.UpdateTaskByID) // good
	protected.DELETE("/accounts/Tasks/:id", task_package.DeleteTaskbyid) // THIS has problem in deleting 
	protected.GET("/accounts/TasksToTODO/:id", task_package.AddTask_TO_TODOMODEL) // THIS has problem in adding

	router.Run(":8081")
}