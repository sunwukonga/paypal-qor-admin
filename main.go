package main

import (
	"fmt"
	"net/http"

	"github.com/sunwukonga/paypal-qor-admin/config"
	"github.com/sunwukonga/paypal-qor-admin/config/admin"
	"github.com/sunwukonga/paypal-qor-admin/config/api"
	_ "github.com/sunwukonga/paypal-qor-admin/config/i18n"
	"github.com/sunwukonga/paypal-qor-admin/config/routes"
	_ "github.com/sunwukonga/paypal-qor-admin/db/migrations"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", routes.Router())
	admin.Admin.MountTo("/admin", mux)
	admin.Widgets.WidgetSettingResource.IndexAttrs("Name")

	api.API.MountTo("/api", mux)
	admin.Filebox.MountTo("/downloads", mux)

	for _, path := range []string{"system", "javascripts", "stylesheets", "images"} {
		mux.Handle(fmt.Sprintf("/%s/", path), http.FileServer(http.Dir("public")))
	}

	fmt.Printf("Listening on: %v\n", config.Config.Port)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", config.Config.Host, config.Config.Port), mux); err != nil {
		panic(err)
	}
}
