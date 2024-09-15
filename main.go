package main

import (
	"log"
	"net/smtp"
	account_package "project/Account"
	task_package "project/Task"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

const (
    smtpServer = "smtp.example.com"
    smtpPort = "587"
    smtpUser = "fatahmed1928@gmail.com"
    smtpPass = "swku eack sobs tvwn"
)

// Function to check tasks and send an email if any task is due within 24 hours
func CheckTasksAndNotify(db *gorm.DB) {
    now := time.Now()
    twentyFourHoursLater := now.Add(24 * time.Hour)

    var tasks []task_package.Tasks
    err := db.Where("end_date BETWEEN ? AND ?", now, twentyFourHoursLater).Find(&tasks).Error
    if err != nil {
        log.Fatalf("Error querying tasks: %v", err)
    }

    if len(tasks) > 0 {
        for _, task := range tasks {
            var account account_package.Account
            err := db.First(&account, "id = ?", task.AccountID).Error
            if err != nil {
                log.Printf("Error retrieving account for task %s: %v", task.TaskID, err)
                continue
            }

            subject := "Upcoming Task Due Soon"
            body := "The following task is due within the next 24 hours:\n" +
                "Task ID: " + task.TaskID + "\n" +
                "Title: " + task.Title + "\n" +
                "End Date: " + task.EndDate.Format(time.RFC1123) + "\n"

            err = SendEmail("fatahmed1928@gmail.com", account.Email, subject, body)
            if err != nil {
                log.Printf("Error sending email for task %s: %v", task.TaskID, err)
            }
        }
    }
}

// Function to send an email
func SendEmail(from, to, subject, body string) error {
    msg := "From: " + from + "\n" +
           "To: " + to + "\n" +
           "Subject: " + subject + "\n\n" +
           body

    auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpServer)
    err := smtp.SendMail(smtpServer+":"+smtpPort, auth, from, []string{to}, []byte(msg))
    return err
}

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

	// Start periodic task checking
    ticker := time.NewTicker(1 * time.Hour) // Check every hour
    go func() {
        for range ticker.C {
            CheckTasksAndNotify(db)
        }
    }()

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


	// Task Manager part

	if err := db.AutoMigrate(&task_package.Tasks{}, &task_package.To_DO_Tasks{}); err != nil {
		panic("failed to migrate database")
	}

	task_package.InitializeDB(db)

	// Create Task
	protected.POST("/accounts/Tasks", task_package.CreateTask) 
	protected.POST("/accounts/Tasks/:accountid", task_package.CreateTaskbyID) 

	// Get Task
	protected.GET("/accounts/Tasks", task_package.GetMyTasks) 
	protected.GET("/accounts/Tasks/:acountid", task_package.GetMyTasksbyID) 

	// Update Task
	protected.PUT("/accounts/Tasks/:taskid", task_package.UpdateMyTask)

	// Delete Task
	protected.DELETE("/accounts/Tasks/:taskid", task_package.DeleteTaskbyid) 

	protected.POST("/accounts/TasksToTODO/:taskid", task_package.AddTask_TO_TODOMODEL) 
	protected.DELETE("/accounts/TasksToTODO/:taskid", task_package.DeleteTODO_TASKbyid) 

	router.Run(":8081")
}