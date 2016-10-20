// +build enterprise

package routes

import "github.com/sunwukonga/qor-example/config/admin"

func init() {
	Router()
	WildcardRouter.AddHandler(admin.MicroSite)
}
