package controllers

import (
	"ambassador/src/database"
	"ambassador/src/models"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"time"
	"context"
	"encoding/json"
	"strings"
	"sort"
)

func Products(c *fiber.Ctx) error {
	var products []models.Product

	database.DB.Find(&products)

	return c.JSON(products)
}

func CreateProduct(c *fiber.Ctx) error {
	var product models.Product

	if err := c.BodyParser(&product); err != nil {
		return err
	}

	database.DB.Create(&product)

	// We clear the redis cache every time we create a product
	go database.ClearCache("products_frontend", "products_backend")

	return c.JSON(product)
}

func GetProduct(c *fiber.Ctx) error {
	var product models.Product

	id, _ := strconv.Atoi(c.Params("id"))


	product.Id = uint(id)

	database.DB.Find(&product)

	return c.JSON(product)
}

func UpdateProduct(c *fiber.Ctx) error {
	
	
	id, _ := strconv.Atoi(c.Params("id"))
	
	
	product := models.Product{}

	product.Id = uint(id)
	
	if err := c.BodyParser(&product); err != nil {
		return err
	}

	// With model we are specifing the table on that we are working
	database.DB.Model(&product).Updates(&product)

	// We remove the cache and rebuild when we update one product
	// database.Cache.Del(context.Background(), "products_backend", "products_frontend")
	// go deleteCache("products_backend") // la palabra reservada "go" es una goroutine y sirve para ejecución asíncrona 
	// go deleteCache("products_frontend")

	// go func(key string) {
	// 	database.Cache.Del(context.Background(), key)
	// }("products_frontend") // esto es como declarar la funcion y ejecutarla directamente
	go database.ClearCache("products_frontend", "products_backend")


	return c.JSON(product)
}

func DeleteProduct(c *fiber.Ctx) error {

	id, _ := strconv.Atoi(c.Params("id"))

	product := models.Product{}

	product.Id = uint(id)

	database.DB.Delete(&product)

	// We clear the redis cache every time we delete a product
	go database.ClearCache("products_frontend", "products_backend")

	return nil
}

func ProductsFrontend(c *fiber.Ctx) error {
	var products []models.Product

	// We cache products provided to the frontend
	var ctx = context.Background()
	// We obtain the value with the key from Redis
	result, err := database.Cache.Get(ctx, "products_frontend").Result()

	if err != nil {
		// this is only executed the first time
		database.DB.Find(&products)
		
		// we have to convert products into bytes
		bytes, err := json.Marshal(products)

		if err != nil {
			panic(err)
		}

		if errKey := database.Cache.Set(ctx, "products_frontend", bytes, 30 * time.Minute).Err(); errKey != nil {
			panic(errKey)
		}
	} else {
		// Here we are returning products cached
		json.Unmarshal([]byte(result), &products)
	}


	return c.JSON(products)
}

func ProductsBackend(c *fiber.Ctx) error {
	var products []models.Product

	// We cache products provided to the backend
	var ctx = context.Background()
	// We obtain the value with the key from Redis
	result, err := database.Cache.Get(ctx, "products_backend").Result()

	if err != nil {
		// this is only executed the first time
		database.DB.Find(&products)
		
		// we have to convert products into bytes
		bytes, err := json.Marshal(products)

		if err != nil {
			panic(err)
		}

		database.Cache.Set(ctx, "products_backend", bytes, 30 * time.Minute)
	} else {
		// Here we are returning products cached
		json.Unmarshal([]byte(result), &products)
	}

	var searchedProducts []models.Product

	if s := c.Query("s"); s != "" {
		lower := strings.ToLower(s)
		for _, product := range products {
			if strings.Contains(strings.ToLower(product.Title), lower) || strings.Contains(strings.ToLower(product.Description), lower) {
				searchedProducts = append(searchedProducts, product)
			}
		}
	} else {
		searchedProducts = products
	}

	if sortParam := c.Query("sort"); sortParam != "" {
		sortLower := strings.ToLower(sortParam)
		if sortLower == "asc" {
			sort.Slice(searchedProducts, func(i, j int) bool {
				return searchedProducts[i].Price < searchedProducts[j].Price
			})
		} else if sortLower == "desc" {
			sort.Slice(searchedProducts, func(i, j int) bool {
				return searchedProducts[i].Price > searchedProducts[j].Price
			})
		}
	}

	var total = len(searchedProducts)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage := 9

	var data []models.Product

	if total <= page * perPage && total >= (page - 1) * perPage {
		data = searchedProducts[(page - 1) * perPage : total]
	} else if total >= page * perPage {
		data = searchedProducts[(page - 1) * perPage : page * perPage]
	} else {
		data = []models.Product{}
	}

	// return c.JSON(data[(page - 1) * perPage : page * perPage])
	return c.JSON(fiber.Map{
		"data": data,
		"total": total,
		"page": page,
		"last_page": total / perPage + 1,
	})
}