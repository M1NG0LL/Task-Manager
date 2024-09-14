package task

import (
	"net/http"

	account_package "project/Account"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitializeDB(database *gorm.DB) {
    db = database
}

func CreateTask(c *gin.Context) {
	accountID, ID_exists := c.Get("accountID")

	if !ID_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var task Tasks
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.TaskID = uuid.New().String()

	task.AccountID = accountID.(string)

	if err := db.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func CreateTaskbyID(c *gin.Context) {
	isAdmin, Admin_exists := c.Get("isAdmin")
	paramID := c.Param("id")

	if !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if !isAdmin.(bool) {
		CreateTask(c)
		return
	}

	var task Tasks
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.TaskID = uuid.New().String()

	task.AccountID = paramID

	if err := db.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func GetMyTasks(c *gin.Context) {
	accountID, ID_exists := c.Get("accountID")
	isAdmin, Admin_exists := c.Get("isAdmin")
	
	if !ID_exists || !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if isAdmin.(bool) {
		GetTasks(c)
		return
	}

	var tasks []Tasks
	if err := db.Where("account_id = ?", accountID).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tasks not found"})
		return
	}

	var toDoTasks []To_DO_Tasks
	if err := db.Where("account_id = ?", accountID).Find(&toDoTasks).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "To-Do tasks not found"})
		return
	}

	var account account_package.Account
	if err := db.First(&account, "id = ?", accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}
	
	output := gin.H{
		"Account": 	  account,
		"Tasks":      tasks,  
		"ToDoTasks":  toDoTasks, 
	}

	c.JSON(http.StatusOK, output)
}

func GetTasks(c *gin.Context) {
	var tasks []Tasks
	var toDoTasks []To_DO_Tasks
	var accounts []account_package.Account

	if err := db.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve tasks"})
		return
	}
	if err := db.Find(&toDoTasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve to-do tasks"})
		return
	}
	if err := db.Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve accounts"})
		return
	}

	output := gin.H{
		"Account": 	  accounts,
		"Tasks":      tasks,
		"ToDoTasks":  toDoTasks,
	}

	c.JSON(http.StatusOK, output)
}

func GetMyTasksbyID(c *gin.Context) {
	isAdmin, Admin_exists := c.Get("isAdmin")
	paramID := c.Param("id")
	
	if !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if !isAdmin.(bool) {
		GetMyTasks(c)
		return
	}

	var tasks []Tasks
	if err := db.Where("account_id = ?", paramID).Find(&tasks).Error; err != nil {  // this is new
		c.JSON(http.StatusNotFound, gin.H{"error": "Tasks not found"})
		return
	}

	var toDoTasks []To_DO_Tasks
	if err := db.Where("account_id = ?", paramID).Find(&toDoTasks).Error; err != nil {  // this is new
		c.JSON(http.StatusNotFound, gin.H{"error": "To-Do tasks not found"})
		return
	}

	var account account_package.Account
	if err := db.First(&account, "id = ?", paramID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	
	output := gin.H{
		"Account": 	  account,
		"Tasks":      tasks,
		"ToDoTasks":  toDoTasks,
	}

	c.JSON(http.StatusOK, output)
}

func UpdateMyTask(c *gin.Context)  {
	accountID, ID_exists := c.Get("accountID")
	isAdmin, Admin_exists := c.Get("isAdmin")

	if !ID_exists || !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var pretask Tasks
	if err := db.First(&pretask, "account_id = ?", accountID).Error; err != nil {  
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var task Tasks
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isAdmin.(bool) {
		task.AccountID = pretask.AccountID
		task.TaskID = pretask.TaskID
	}

	if err := db.Model(&pretask).Updates(task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Task"})
		return
	}

	var toDoTasks To_DO_Tasks
	if err := db.Model(&pretask).Updates(toDoTasks).Error; err != nil {  
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update To-Do task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func UpdateTaskByID(c *gin.Context) {
	isAdmin, Admin_exists := c.Get("isAdmin")

	if  !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	paramID := c.Param("id")

	if !isAdmin.(bool) {
		c.Set("accountID", paramID) 
		UpdateMyTask(c)
		return
	} 
	
	var task Tasks
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&Tasks{}).Where("account_id = ?", paramID).Updates(task).Error; err != nil { 
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	var toDoTasks To_DO_Tasks
	if err := db.Model(&To_DO_Tasks{}).Updates(toDoTasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update To-Do task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully"})
}

func DeleteTaskbyid(c *gin.Context)  {
	isAdmin, Admin_exists := c.Get("isAdmin")

	if !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	paramID := c.Param("id")

	if !isAdmin.(bool) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This url is for ADMIN ONLY."})
		return
	}

	var task Tasks
	if err := db.First(&task, "id = ?", paramID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var to_do_task To_DO_Tasks
	if err := db.First(&to_do_task, "id = ?", paramID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if err := db.Delete(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete Task"})
		return
	}
	if err := db.Delete(&to_do_task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete Task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

func AddTask_TO_TODOMODEL(c *gin.Context)  {
	_, ID_exists := c.Get("accountID")
	paramID := c.Param("id")

	if !ID_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var task Tasks
	if err := db.First(&task, "id = ?", paramID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var to_do_task To_DO_Tasks

	if err := db.Model(&task).Create(to_do_task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task has been added to TO DO TASKS"})
}