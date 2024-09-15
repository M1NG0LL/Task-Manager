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

// POST
// Create task by the user
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

// POST
// Create Task by Admin
func CreateTaskbyID(c *gin.Context) {
	isAdmin, Admin_exists := c.Get("isAdmin")
	accountID := c.Param("accountid")

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

	task.AccountID = accountID

	if err := db.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// GET
// Get all Tasks and TODO TASKS of the users (Admin function)
func GetTasks(c *gin.Context) {
    var accounts []account_package.Account
    if err := db.Find(&accounts).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve accounts"})
        return
    }

    output := []gin.H{}

    for _, account := range accounts {
        var tasks []Tasks
        var toDoTasks []To_DO_Tasks

        // Find tasks and to-do tasks associated with this account
        if err := db.Where("account_id = ?", account.ID).Find(&tasks).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve tasks for account"})
            return
        }

        if err := db.Where("account_id = ?", account.ID).Find(&toDoTasks).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve to-do tasks for account"})
            return
        }

        // Append account with associated tasks and to-do tasks to output
        accountData := gin.H{
            "Account":    account,
            "Tasks":      tasks,
            "ToDoTasks":  toDoTasks,
        }

        output = append(output, accountData)
    }

    c.JSON(http.StatusOK, gin.H{"Accounts": output})
}

// GET
// Get all Tasks and TODO TASKS of the user
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

// GET
// Get all Tasks and TODO TASKS of the user (Admin function)
func GetMyTasksbyID(c *gin.Context) {
	isAdmin, Admin_exists := c.Get("isAdmin")
	accountID := c.Param("accountid")
	
	if !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if !isAdmin.(bool) {
		GetMyTasks(c)
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

// PUT
// Update Task(if this task was in TODO model it will be updated too) by user 
// This Function is for both the user and Admin
func UpdateMyTask(c *gin.Context) {
	accountID, ID_exists := c.Get("accountID")
	isAdmin, Admin_exists := c.Get("isAdmin")
	taskID := c.Param("taskid")

	if !ID_exists || !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var pretask Tasks
	if err := db.First(&pretask, "task_id = ?", taskID).Error; err != nil {  
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var task Tasks
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isAdmin.(bool) {
        if pretask.AccountID != accountID {
            c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this task"})
            return
        }
        task.AccountID = pretask.AccountID 
        task.TaskID = pretask.TaskID       
    }

	if err := db.Model(&pretask).Updates(task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Task"})
		return
	}

	var existingToDoTask To_DO_Tasks
    if err := db.First(&existingToDoTask, "task_id = ?", taskID).Error; err == nil {
        toDoTaskUpdates := map[string]interface{}{
			"TaskID": 	   task.TaskID,
			"AccountID":   task.AccountID,
            "Title":       task.Title,
            "Description": task.Description,
            "StartDate":   task.StartDate,
            "EndDate":     task.EndDate,
        }
        if err := db.Model(&existingToDoTask).Updates(toDoTaskUpdates).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update To-Do task"})
            return
        }
    }

	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully", "task": task})
}

// DELETE
// Delete Task (if this task was in TODO model it will be deleted too) by Admin (Admin function)
func DeleteTaskbyid(c *gin.Context) {
	isAdmin, Admin_exists := c.Get("isAdmin")

	if !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	if !isAdmin.(bool) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This URL is for ADMIN ONLY."})
		return
	}

	taskID := c.Param("taskid")

	var task Tasks
	if err := db.First(&task, "task_id = ?", taskID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding Task"})
		}
		return
	}

	if err := db.Delete(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete Task"})
		return
	}

	var toDoTask To_DO_Tasks
	if err := db.First(&toDoTask, "task_id = ?", taskID).Error; err == nil {
		if err := db.Delete(&toDoTask).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete To-Do task"})
			return
		}
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding To-Do task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task and To-Do task deleted successfully"})
}

// POST
// Add Task from TASKS model to TODO model
func AddTask_TO_TODOMODEL(c *gin.Context) {
    _, ID_exists := c.Get("accountID")
    taskID := c.Param("taskid")

    if !ID_exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    var task Tasks
    if err := db.First(&task, "task_id = ?", taskID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
        return
    }

    var existingTask To_DO_Tasks
    if err := db.First(&existingTask, "task_id = ?", taskID).Error; err == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Task already exists in To Do Tasks"})
        return
    } else if err != gorm.ErrRecordNotFound {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking To-Do Tasks"})
        return
    }

    toDoTask := To_DO_Tasks(task)

    if err := db.Create(&toDoTask).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Task has been added to TO DO TASKS"})
}

// DELETE
// Delete TODO task from TODO model 
func DeleteTODO_TASKbyid(c *gin.Context) {
	_, ID_exists := c.Get("accountID")
    taskID := c.Param("taskid")

	if !ID_exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

	var toDoTask To_DO_Tasks
	if err := db.First(&toDoTask, "task_id = ?", taskID).Error; err == nil {
		if err := db.Delete(&toDoTask).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete To-Do task"})
			return
		}
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding To-Do task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "To-Do task deleted successfully"})
}