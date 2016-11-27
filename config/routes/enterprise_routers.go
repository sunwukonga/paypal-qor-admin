// +build enterprise

package routes

import "github.com/sunwukonga/paypal-qor-admin/config/admin"

func init() {
	Router()
	WildcardRouter.AddHandler(admin.MicroSite)
}
