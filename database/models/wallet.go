package models

import "github.com/google/uuid"

type Wallet struct {
	ID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	UserID uuid.UUID `gorm:"type:uuid;not null;OnDelete:CASCADE"`
	Coin   uint      `gorm:"not null;default:1000;check:Coin >= 0"`
}
