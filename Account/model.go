package account_package

type Account struct {
	ID        string `gorm:"primaryKey"`
	FirstName string
	LastName  string
	Username  string `gorm:"unique;not null"`
	Email     string `gorm:"unique;not null"`
	Password  string
	IsActive  bool `gorm:"default:false"`
	IsAdmin   bool `gorm:"default:false"`
}
