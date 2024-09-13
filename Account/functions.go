package account_package

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var db *gorm.DB
var jwtKey = []byte("secret_key")

// JWT Claims structure
type Claims struct {
	ID        string `gorm:"primaryKey"`
	IsActive  bool `gorm:"default:false"`
	IsAdmin   bool `gorm:"default:false"`

	jwt.StandardClaims
}

func Init(database *gorm.DB) {
	db = database
}

func Login(c *gin.Context) {
	var loginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var account Account
	if err := db.Where("username = ?", loginRequest.Username).First(&account).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if account.Password != loginRequest.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if !account.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account not active"})
		return
	}

	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		ID:       account.ID,
		IsActive: account.IsActive,
		IsAdmin:  account.IsAdmin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}


// Middleware to verify JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the Authorization header
		tokenString := c.GetHeader("Authorization")
		if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header not provided or malformed"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix from the token string
		tokenString = tokenString[7:]

		// Parse the token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		// Check if the token is valid
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set the account details into the context for further use
		c.Set("accountID", claims.ID)
		c.Set("isActive", claims.IsActive)
		c.Set("isAdmin", claims.IsAdmin)

		// Proceed to the next handler
		c.Next()
	}
}

// POST
// Create account
func CreateAccount(c *gin.Context) {
	var account Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingAccount Account
	if err := db.Where("username = ? OR email = ?", account.Username, account.Email).First(&existingAccount).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username or Email already exists"})
		return
	}

	namespace := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8") 
	account.ID = uuid.NewMD5(namespace, []byte(account.Username)).String()

	account.IsActive = false
	account.IsAdmin = false

	if err := db.Create(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, account)
}

// PUT
// To activate the account
func ActivateAccountByID(c *gin.Context) {
	accountID := c.Param("id")

	var account Account
	if err := db.First(&account, "id = ?", accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	account.IsActive = true
	if err := db.Save(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account activated successfully"})
}

// GET
// get Account info using the Token
func GetMyAccount(c *gin.Context) {
	accountID, ID_exists := c.Get("accountID")
	isAdmin, Admin_exists := c.Get("isAdmin")
	
	if !ID_exists || !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if isAdmin.(bool) {
		GetAccounts(c)
		return
	}

	var account Account
	if err := db.First(&account, "id = ?", accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}
	
	c.JSON(http.StatusOK, account)
}

// get Accounts if you are admin using the same url of getMyAccount()
func GetAccounts(c *gin.Context) {
	var accounts []Account
	if err := db.Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve accounts"})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// PUT
// Update Account info using the Token
func UpdateMyAccount(c *gin.Context) {
	accountID, ID_exists := c.Get("accountID")
	isAdmin, Admin_exists := c.Get("isAdmin")

	if !ID_exists || !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var preaccount Account
	if err := db.First(&preaccount, "id = ?", accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	var account Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isAdmin.(bool) {
		account.ID = preaccount.ID
		account.IsActive = preaccount.IsActive
		account.IsAdmin = preaccount.IsAdmin
	}

	if err := db.Model(&preaccount).Updates(account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account"})
		return
	}

	c.JSON(http.StatusOK, account)
}

// PUT
// This func is for ADMINS ONLY
// Update any account by putting id in url 
func UpdateAccountByID(c *gin.Context) {
	isAdmin, Admin_exists := c.Get("isAdmin")

	if  !Admin_exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	paramID := c.Param("id")

	if !isAdmin.(bool) {
		c.Set("accountID", paramID) 
		UpdateMyAccount(c)
		return
	} 
	
	var account Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&Account{}).Where("id = ?", paramID).Updates(account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account updated successfully"})
}

// DELETE
// This func is for ADMINS ONLY
// Delete any account by putting id in url 
func DeleteAccountbyid(c *gin.Context)  {
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

	var account Account
	if err := db.First(&account, "id = ?", paramID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	if err := db.Delete(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}