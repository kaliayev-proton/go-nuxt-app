package models

type Product struct {
	Model // here Product is inheriting Model fields
	Title string `json:"title"`
	Description string `json:"description"`
	Image string `json:"image"`
	Price float64 `json:"price"`
}