package main

import (
	"fmt"
	"net/http"
	//	"strings"

	"github.com/gorilla/csrf"
	"github.com/qor/qor/utils"
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
		mux.Handle(fmt.Sprintf("/%s/", path), utils.FileServer(http.Dir("public")))
	}

	/*	for _, path := range []string{"system", "javascripts", "stylesheets"} {
		publicDir := http.Dir(strings.Join([]string{config.Root, "public"}, "/"))
		mux.Handle(fmt.Sprintf("/%s/", path), http.FileServer(publicDir))
	} */

	fmt.Printf("Listening on: %v\n", config.Config.Port)

	skipCheck := func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/auth") {
				r = csrf.UnsafeSkipCheck(r)
			}
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
	handler := csrf.Protect([]byte("3693f371bf91487c99286a777811bd4e"), csrf.Secure(false))(mux)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", config.Config.Host, config.Config.Port), skipCheck(handler)); err != nil {
		panic(err)
	}
}
