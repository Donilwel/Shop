package models

import (
	"github.com/google/uuid"
	"time"
)

// Purchase
//
// @Description Структура сделки
type Purchase struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;OnDelete:CASCADE"`
	MerchID   uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt time.Time `gorm:"precision:6"`
}
