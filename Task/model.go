package task

type Tasks struct {
	TaskID      string `gorm:"primaryKey"`
	AccountID   string
	Title       string `gorm:"not null"`
	Description string
	StartDate   string `gorm:"not null"`
	EndDate     string `gorm:"not null"`
}

type To_DO_Tasks struct {
	TaskID      string `gorm:"primaryKey"`
	AccountID   string
	Title       string `gorm:"not null"`
	Description string
	StartDate   string `gorm:"not null"`
	EndDate     string `gorm:"not null"`
}
