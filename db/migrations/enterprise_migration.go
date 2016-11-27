// +build enterprise

package migrations

import "github.com/sunwukonga/paypal-qor-admin/config/admin"

func init() {
	AutoMigrate(&admin.QorMicroSite{})
}
