// +build enterprise

package migrations

import "github.com/sunwukonga/qor-example/config/admin"

func init() {
	AutoMigrate(&admin.QorMicroSite{})
}
