package controllers

import (
	"ambassador/src/database"
	"ambassador/src/models"
	"github.com/gofiber/fiber/v2"
)

func Orders(c *fiber.Ctx) error {
	var orders []models.Order

	database.DB.Preload("OrderItems").Find(&orders)

	// We create the field name to show it in the response
	for i, order := range orders {
		orders[i].Name = order.FullName()
		orders[i].Total = order.GetTotal()
	}

	return c.JSON(orders)
}