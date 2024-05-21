package models

import "gorm.io/gorm"

type Customer struct {
	gorm.Model
	Name    string `json:"name" gorm:"text;not null;default:null`
	Email   string `json:"email" gorm:"primaryKey;type:varchar(100);not null;uniqueIndex"`
	Address string `json:"address" gorm:"text;not null;default:null`
	Number  int    `json:"number" gorm:"not null;default:0`
}
