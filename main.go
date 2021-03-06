package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/csrf"
	qoradmin "github.com/qor/admin"
	"github.com/qor/qor/utils"
	"github.com/sunwukonga/paypal-qor-admin/app/controllers"
	"github.com/sunwukonga/paypal-qor-admin/app/models"
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

	admin.Admin.GetRouter().Use(&qoradmin.Middleware{
		Name: "influencerAuth",
		Handler: func(context *qoradmin.Context, middleware *qoradmin.Middleware) {
			user := context.CurrentUser.(*models.User)
			if user.Role != models.RoleInfluencer {
				middleware.Next(context)
				return
			} else {
				influencerCoupon := &models.InfluencerCoupon{}
				if err := context.DB.Where("user_id = ?", user.ID).First(influencerCoupon).Error; err == nil {
					if influencerCoupon.Active {
						middleware.Next(context)
						return
					}
				}
				// redirect to box order route.
				http.Redirect(context.Writer, context.Request, "/influencer/buysamplebox", http.StatusSeeOther)
				return
			}
		},
	})
	admin.Admin.MountTo("/admin", mux)
	admin.Widgets.WidgetSettingResource.IndexAttrs("Name")

	api.API.MountTo("/api", mux)
	admin.Filebox.MountTo("/downloads", mux)

	// Open IPN log file and create the logger.
	f, err := os.OpenFile("ipn.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
	}
	defer f.Close()
	controllers.IPNLogger = log.New(f, "", log.LstdFlags)

	/*
		for _, path := range []string{"system", "javascripts", "stylesheets", "images"} {
			mux.Handle(fmt.Sprintf("/%s/", path), utils.FileServer(http.Dir("public")))
		}
	*/
	for _, path := range []string{"system", "javascripts", "stylesheets"} {
		publicDir := http.Dir(strings.Join([]string{config.Root, "public"}, "/"))
		mux.Handle(fmt.Sprintf("/%s/", path), utils.FileServer(publicDir))
	}

	fmt.Printf("Root dir: %v\n", strings.Join([]string{config.Root, "public"}, "/"))
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
