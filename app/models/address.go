package models

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

const (
	DescriptionPaypal   string = "Registered Paypal Address"
	DescriptionDelivery string = "Delivery Address"
)

type Address struct {
	gorm.Model
	UserID      uint
	Description string
	ContactName string
	Phone       string
	Country     string
	City        string
	Address1    string
	Address2    string
	Postcode    string
}

func (address Address) Stringify() string {
	return fmt.Sprintf("%v, %v, %v", address.Address2, address.Address1, address.City)
}
