package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/qor/seo"
	"github.com/qor/widget"
	"github.com/sunwukonga/paypal-qor-admin/app/models"
	"github.com/sunwukonga/paypal-qor-admin/config"
	"github.com/sunwukonga/paypal-qor-admin/config/admin"
	"github.com/sunwukonga/paypal-qor-admin/config/auth"
	"gopkg.in/authboss.v0"
)

func HomeIndex(ctx *gin.Context) {
	var products []models.Product
	//TODO: Put information into User struct.
	//b, _ := json.Marshal(User{})
	// Convert bytes to string.
	//cartString := string(b)
	//sessionStore.Put("session_cart", cartString)

	DB(ctx).Limit(9).Preload("ColorVariations").Find(&products)
	seoObj := models.SEOSetting{}
	DB(ctx).First(&seoObj)

	widgetContext := admin.Widgets.NewContext(&widget.Context{
		DB:         DB(ctx),
		Options:    map[string]interface{}{"Request": ctx.Request},
		InlineEdit: IsEditMode(ctx),
	})

	config.View.Funcs(I18nFuncMap(ctx)).Execute(
		"home_index",
		gin.H{
			"ActionBarTag":           admin.ActionBar.Render(ctx.Writer, ctx.Request),
			authboss.FlashSuccessKey: auth.Auth.FlashSuccess(ctx.Writer, ctx.Request),
			authboss.FlashErrorKey:   auth.Auth.FlashError(ctx.Writer, ctx.Request),
			"SeoTag":                 seoObj.HomePage.Render(seoObj, nil),
			"top_banner":             widgetContext.Render("TopBanner", "Banner"),
			"feature_products":       widgetContext.Render("FeatureProducts", "Products"),
			"Products":               products,
			"MicroSearch": seo.MicroSearch{
				URL:    "http://demo.getqor.com",
				Target: "http://demo.getqor.com/search?q={keyword}",
			}.Render(),
			"MicroContact": seo.MicroContact{
				URL:         "http://demo.getqor.com",
				Telephone:   "080-0012-3232",
				ContactType: "Customer Service",
			}.Render(),
			"CurrentUser":   CurrentUser(ctx),
			"CurrentLocale": CurrentLocale(ctx),
		},
		ctx.Request,
		ctx.Writer,
	)
}
