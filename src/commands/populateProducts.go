package main

import (
	"github.com/bxcodec/faker/v3"
	"ambassador/src/database"
	"ambassador/src/models"
	"math/rand"
)

// IMPORTANTE: Para ejecutar estos comandos tenemos que estar dentro del contendor de Docker

func main() {
	database.Connect()
	for i := 0; i < 30; i++ {
		product := models.Product{
			Title: faker.Username(),
			Description: faker.Username(),
			Image: faker.URL(),
			Price: float64(rand.Intn(90) + 10),
		}

		database.DB.Create(&product)
	}
}

