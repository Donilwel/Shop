package models

import (
	"github.com/google/uuid"
	"time"
)

type Transaction struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	FromUser  uuid.UUID `gorm:"type:uuid;not null"`
	ToUser    uuid.UUID `gorm:"type:uuid;not null"`
	Amount    uint      `gorm:"not null"`
	CreatedAt time.Time `gorm:"precision:6"`
}
