package models

import (
	"github.com/google/uuid"
	"time"
)

const (
	ADMIN_ROLE    string = "ADMIN_ROLE"
	EMPLOYEE_ROLE string = "EMPLOYEE_ROLE"
)

// User
//
// @Description Структура user
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Username  string    `gorm:"type:varchar(100);unique;not null"`
	Email     string    `gorm:"type:varchar(100);unique;not null"`
	Password  string    `gorm:"type:varchar(255);not null"`
	Role      string    `gorm:"type:varchar(100);not null;default:'EMPLOYEE_ROLE'"`
	CreatedAt time.Time `gorm:"precision:6"`
	UpdatedAt time.Time `gorm:"precision:6"`
}
