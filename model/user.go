package model

import "time"

type User struct {
	ID        string `gorm:"primaryKey"`
	Email     string `gorm:"unique"`
	Password  string
	LastLogin time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
