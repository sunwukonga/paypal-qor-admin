package admin

import (
	"github.com/qor/filebox"
	"github.com/sunwukonga/paypal-qor-admin/config"
	"github.com/sunwukonga/paypal-qor-admin/config/auth"
	"github.com/qor/roles"
)

var Filebox *filebox.Filebox

func init() {
	Filebox = filebox.New(config.Root + "/public/downloads")
	Filebox.SetAuth(auth.AdminAuth{})
	dir := Filebox.AccessDir("/")
	dir.SetPermission(roles.Allow(roles.Read, "admin"))
}
