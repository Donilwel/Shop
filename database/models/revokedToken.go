package models

import (
	"github.com/google/uuid"
	"time"
)

type RevokedToken struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	Token     string    `gorm:"unique;not null"`
	CreatedAt time.Time
}
