package models

import (
	"database/sql"
	"time"

	"github.com/jinzhu/gorm"
)

const (
	RoleAdmin      string = "Admin"
	RoleCustomer   string = "Customer"
	RoleReseller   string = "Reseller"
	RoleInfluencer string = "Influencer"
	RoleSubscriber string = "Subscriber"
	RoleEditor     string = "Editor"
	RoleServicer   string = "Servicer"
)

var Roles = []string{RoleAdmin, RoleCustomer, RoleReseller, RoleInfluencer, RoleSubscriber, RoleEditor, RoleServicer}

type User struct {
	gorm.Model
	Email     string
	Password  string
	Name      sql.NullString
	Gender    string
	Role      string
	Addresses []Address

	// Confirm
	ConfirmToken string
	Confirmed    bool

	// Recover
	RecoverToken       string
	RecoverTokenExpiry *time.Time
}

type InfluencerCoupon struct {
	gorm.Model
	Code   string
	User   User
	UserID uint
}

func (user User) DisplayName() string {
	return user.Email
}

func (user User) AvailableLocales() []string {
	return []string{"en-US", "zh-CN"}
}
