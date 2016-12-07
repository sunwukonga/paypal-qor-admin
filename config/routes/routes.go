package routes

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	//	"github.com/qor/log"
	"github.com/qor/qor"
	"github.com/qor/qor/utils"
	"github.com/qor/wildcard_router"
	"github.com/sunwukonga/paypal-qor-admin/app/controllers"
	"github.com/sunwukonga/paypal-qor-admin/config"
	"github.com/sunwukonga/paypal-qor-admin/config/auth"
	"github.com/sunwukonga/paypal-qor-admin/db"
)

var rootMux *http.ServeMux
var WildcardRouter *wildcard_router.WildcardRouter

func Router() *http.ServeMux {
	if rootMux == nil {
		router := gin.Default()
		router.Use(func(ctx *gin.Context) {
			if locale := utils.GetLocale(&qor.Context{Request: ctx.Request, Writer: ctx.Writer}); locale != "" {
				ctx.Set("DB", db.DB.Set("l10n:locale", locale))
			}
		})
		//		router.Use(log.Logger("application.log", 30))
		gin.SetMode(gin.DebugMode)

		//router.GET("/", controllers.HomeIndex)
		router.GET("/", func(c *gin.Context) {
			c.Redirect(302, "/admin")
		})
		router.GET("/products/:code", controllers.ProductShow)
		router.GET("/cart", controllers.CartShow)
		router.GET("/switch_locale", controllers.SwitchLocale)

		//TODO: Fix this. Should be a POST route.
		router.POST("/addtocart/:id", controllers.AddToCart)
		router.POST("/deletefromcart/:id", controllers.DeleteFromCart)
		router.POST("/updateitemqty/:id", controllers.UpdateItemQty)

		//Checking route for influencer codes, returns simple true or false
		router.GET("/couponcode/:code", controllers.CodeExists)
		router.POST("/notify", controllers.IpnReceiver)

		rootMux = http.NewServeMux()
		rootMux.Handle("/auth/", auth.Auth.NewRouter())
		publicDir := http.Dir(strings.Join([]string{config.Root, "public"}, "/"))
		rootMux.Handle("/dist/", utils.FileServer(publicDir))
		rootMux.Handle("/vendors/", utils.FileServer(publicDir))
		rootMux.Handle("/images/", utils.FileServer(publicDir))
		rootMux.Handle("/fonts/", utils.FileServer(publicDir))

		WildcardRouter = wildcard_router.New()
		WildcardRouter.MountTo("/", rootMux)
		WildcardRouter.AddHandler(router)
	}
	return rootMux
}
