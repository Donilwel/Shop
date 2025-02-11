package utils

import (
	"Shop/database/migrations"
	"Shop/database/models"
	"fmt"
	"math/rand"
	"time"
)

var adjectives = []string{"Fast", "Crazy", "Cool", "Brave", "Smart", "Lucky", "Wild", "Slowed", "Bad", "Good", "Sick", "Punished", "Elite", "Sweet"}
var nouns = []string{"Tiger", "Eagle", "Wolf", "Shark", "Panther", "Hawk", "Dragon", "Chicken", "Pow", "Dog", "Cat", "Pig", "Lion"}

func GenerateUsername() string {
	rand.Seed(time.Now().UnixNano())
	for {
		adj := adjectives[rand.Intn(len(adjectives))]
		noun := nouns[rand.Intn(len(nouns))]
		number := rand.Intn(9000) + 1000
		username := fmt.Sprintf("%s%s%d", adj, noun, number)

		var count int64
		migrations.DB.Model(&models.User{}).Where("username = ?", username).Count(&count)
		if count == 0 {
			return username
		}
	}
}
