package task

import "time"

type Tasks struct {
	TaskID      string `gorm:"primaryKey"`
	AccountID   string
	Title       string `gorm:"not null"`
	Description string
	StartDate   time.Time `gorm:"not null"`
	EndDate     time.Time `gorm:"not null"`
}

type To_DO_Tasks struct {
	TaskID      string `gorm:"primaryKey"`
	AccountID   string
	Title       string `gorm:"not null"`
	Description string
	StartDate   time.Time `gorm:"not null"`
	EndDate     time.Time `gorm:"not null"`
}
