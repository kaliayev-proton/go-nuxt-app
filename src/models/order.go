package models

type Order struct {
	Model
	TransactionId string `json:"transaction_id" gorm:"null"`
	UserId uint `json:"user_id"`
	Code string `json:"code"`
	AmbassadorEmail string `json:"ambassador_email"`
	FirstName string `json:"-"`
	// we don't show this field in the response
	LastName string `json:"-"`
	// we wont add this field into the database
	Name string `json:"name" gorm:"-"`
	Email string `json:"email"`
	Address string `json:"address" gorm:"null"`
	City string `json:"city" gorm:"null"`
	Country string `json:"country" gorm:"null"`
	Zip string `json:"zip" gorm:"null"`
	Complete bool `json:"-" gorm:"default:false"`
	Total float64 `json:"total" gorm:"-"`
	OrderItems []OrderItem `json:"order_items" gorm:"foreignKey:OrderId"`
}

type OrderItem struct {
	Model
	OrderId uint `json:"order_id"`
	ProductTitle string `json:"order_title"`
	Price float64 `json:"order_id"`
	Quantity uint `json:"quantity"`
	AdminRevenue float64 `json:"admin_revenue"`
	AmbassadorRevenue float64 `json:"ambassador_revenue"`
}

func (order *Order) FullName() string {
	return order.FirstName + " " + order.LastName
}
func (order *Order) GetTotal() float64 {
	var total float64 = 0

	for _, orderItem := range order.OrderItems {
		total += orderItem.Price * float64(orderItem.Quantity)
	}

	return total
}