package main

import (
	"github.com/bxcodec/faker/v3"
	"ambassador/src/database"
	"ambassador/src/models"
)

// IMPORTANTE: Para ejecutar estos comandos tenemos que estar dentro del contendor de Docker

func main() {
	database.Connect()
	for i := 0; i < 30; i++ {
		ambassador := models.User{
			FirstName: faker.FirstName(),
			LastName: faker.LastName(),
			Email: faker.Email(),
			IsAmbassador: true,
		}

		ambassador.SetPassword("12345")

		database.DB.Create(&ambassador)
	}
}

