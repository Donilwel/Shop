package models

import "github.com/google/uuid"

// Merch
//
// @Description Структура сделки
type Merch struct {
	ID    uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	Name  string    `gorm:"unique;not null"`
	Price uint      `gorm:"not null"`
}
