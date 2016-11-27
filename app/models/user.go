package models

import (
	"database/sql"
	"time"

	"github.com/jinzhu/gorm"
)

const (
	RoleAdmin      = "Admin"
	RoleCustomer   = "Customer"
	RoleReseller   = "Reseller"
	RoleInfluencer = "Influencer"
	RoleEditor     = "Editor"
	RoleServicer   = "Servicer"
)

type User struct {
	gorm.Model
	Email     string
	Password  string
	Name      sql.NullString
	Gender    string
	Role      string
	Addresses []Address

	InfluencerCode string

	// Confirm
	ConfirmToken string
	Confirmed    bool

	// Recover
	RecoverToken       string
	RecoverTokenExpiry *time.Time
}

/*type InfluencerCode struct {
	gorm.Model
	Code   string
	User   User
	UserID int
}*/

func (user User) DisplayName() string {
	return user.Email
}

func (user User) AvailableLocales() []string {
	return []string{"en-US", "zh-CN"}
}
