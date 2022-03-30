package controllers

import (
	"ambassador/src/database"
	"ambassador/src/models"
	"ambassador/src/middlewares"
	"github.com/gofiber/fiber/v2"
	"github.com/bxcodec/faker/v3"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"strconv"
	"context"
	"fmt"
	"net/smtp"
)

func Link(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	var links []models.Link
	database.DB.Where("user_id = ?", id).Find(&links)

	for i, link := range links {
		var orders []models.Order
		database.DB.Where("code = ? and complete = true", link.Code).Find(&orders)

		links[i].Orders = orders
	}

	return c.JSON(links)
}

type CreateLinkRequest struct {
	Products []int
}

func CreateLink(c *fiber.Ctx) error {
	var request CreateLinkRequest

	if err := c.BodyParser(&request); err != nil {
		return err
	}

	id, _ := middlewares.GetUserId(c)

	link := models.Link{
		UserId: id,
		Code: faker.Username(),
	}

	for _, productId := range request.Products {
		product := models.Product{}
		product.Id = uint(productId)
		link.Products = append(link.Products, product)
	}

	database.DB.Create(&link)

	return c.JSON(link)
}

// devuelve el revenue por cada link
func Stats(c *fiber.Ctx) error {
	id, _ := middlewares.GetUserId(c)

	var links []models.Link

	database.DB.Find(&links, models.Link{
		UserId: id,
	})

	var result []interface{}

	var orders []models.Order

	for _, link := range links {
		database.DB.Preload("OrderItems").Find(&orders, &models.Order{
			Code: link.Code,
			Complete: true,
		})

		revenue := 0.0

		for _, order := range orders {
			revenue += order.GetTotal()
		}

		result = append(result, fiber.Map{
			"code": link.Code,
			"count": len(orders),
			"revenue": revenue,
		})
	}

	return c.JSON(result)
}

func GetLink(c *fiber.Ctx) error {
	code := c.Params("code")

	link := models.Link{
		Code: code,
	}

	database.DB.Preload("User").Preload("Products").First(&link)

	return c.JSON(link)
}

type CreateOrderRequest struct {
	Code string
	FirstName string
	LastName string
	Email string
	Address string
	Country string
	City string
	Zip string
	Products []map[string]int
}

func CreateOrder(c *fiber.Ctx) error {
	var request CreateOrderRequest

	if err := c.BodyParser(&request); err != nil {
		return err
	}

	link := models.Link{
		Code: request.Code,
	}

	database.DB.Preload("User").First(&link)

	if link.Id == 0 {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Invalid Link",
		})
	}

	order := models.Order{
		Code: link.Code,
		UserId: link.UserId,
		AmbassadorEmail: link.User.Email,
		FirstName: request.FirstName,
		LastName: request.LastName,
		Email: request.Email,
		Address: request.Address,
		Country: request.Country,
		City: request.City,
		Zip: request.Zip,
	}

	// TRANSACTIONS WITH GO
	tx := database.DB.Begin()

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	var lineItems []*stripe.CheckoutSessionLineItemParams

	for _, requestProduct := range request.Products {
		product := models.Product{}
		product.Id = uint(requestProduct["product_id"])
		database.DB.First(&product)

		total := product.Price * float64(requestProduct["quantity"])

		item := models.OrderItem{
			OrderId: order.Id,
			ProductTitle: product.Title,
			Price: product.Price,
			Quantity: uint(requestProduct["quantity"]),
			AmbassadorRevenue: 0.1 * total,
			AdminRevenue: 0.9 * total,
		}

		if err := tx.Create(&item).Error; err != nil {
			tx.Rollback()
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			Name: stripe.String(product.Title),
			Description: stripe.String(product.Description),
			Images: []*string{stripe.String(product.Image)},
			Amount: stripe.Int64(100 * int64(product.Price)),
			Currency: stripe.String("usd"),
			Quantity: stripe.Int64(int64(requestProduct["quantity"])),
		})
	}

	stripe.Key = "sk_test_51KiklmEZPy5gPipJH2bxKHRnNqKtJeMMkYyCXtv1AxlT6Lg3obFcxGiQ7LXp4L1ap0NacWmikPonAuI75m1URNml00mGoVGSSe"

	params := stripe.CheckoutSessionParams{
		SuccessURL: stripe.String("http://localhost:5000/success={CHECKOUT_SESSION_ID}"),
		CancelURL: stripe.String("http://localhost:5000/error"),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: lineItems,
	}

	source, err := session.New(&params)

	if err != nil {
		tx.Rollback()
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	order.TransactionId = source.ID

	// here we save because we just change one value during the transaction
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	tx.Commit()

	return c.JSON(source)
}

func CompleteOrder(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	order := models.Order{}

	database.DB.Preload("OrderItems").First(&order, models.Order{
		TransactionId: data["source"],
	})

	if order.Id == 0 {
		c.Status(fiber.StatusNotFound)
		return c.JSON(fiber.Map{
			"message": "Order not found",
		})
	}

	order.Complete = true
	database.DB.Save(&order)

	go func(order models.Order) {
		ambassadorRevenue := 0.0
		adminRevenue := 0.0
		for _, item := range order.OrderItems {
			ambassadorRevenue += item.AmbassadorRevenue
			adminRevenue += item.AdminRevenue
		}

		user := models.User{}
		user.Id = order.UserId

		database.DB.First(&user)

		database.Cache.ZIncrBy(context.Background(), "rankings", ambassadorRevenue, user.Name())

		// SEND EMAIL to AMBASSADOR
		ambassadorMessage := []byte(fmt.Sprintf("You earned $%f from the link #%s", ambassadorRevenue, order.Code))
		smtp.SendMail("0.0.0.0:1025", nil, "no-replay@email.com", []string{order.AmbassadorEmail}, ambassadorMessage)

		// SEND EMAIL to ADMIN
		adminMessage := []byte(fmt.Sprintf("Order #%d with a total of $%f has been completed", order.Id, adminRevenue))
		// with host.docker.internal:1025 docker cis able to access our localhost
		smtp.SendMail("host.docker.internal:1025", nil, "no-replay@email.com", []string{"admin@admin.com"}, adminMessage)
	}(order)

	return c.JSON(fiber.Map{
		"message": "success",
	})
}