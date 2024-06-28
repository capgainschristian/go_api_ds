package models

import "gorm.io/gorm"

type Customer struct {
	gorm.Model
	Name    string `json:"name" gorm:"type:text;not null;default:null"`
	Email   string `json:"email" gorm:"primaryKey;type:varchar(100);not null;uniqueIndex"`
	Address string `json:"address" gorm:"type:text;not null;default:null"`
	Number  int    `json:"number" gorm:"not null;default:0"`
}

type User struct {
	gorm.Model
	Email   string `json:"email" gorm:"primaryKey;type:varchar(100);not null;uniqueIndex"`
	Password string  `json:"password" gorm:"type:text;not null;default:null"`
}