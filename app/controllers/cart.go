package controllers

import (
	//	"encoding/json"
	//	"html/template"
	//	"strings"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sunwukonga/paypal-qor-admin/app/models"
	"github.com/sunwukonga/paypal-qor-admin/config"
	"github.com/sunwukonga/paypal-qor-admin/config/admin"
	//	"github.com/sunwukonga/paypal-qor-admin/config/auth" // for sessionStore
)

func CartShow(ctx *gin.Context) {
	var (
		cart *models.Order
		//sessionStorer  *auth.SessionStorer
		colorVariation models.ColorVariation
		user           *models.User
	)

	user = CurrentUser(ctx)
	cart = &models.Order{}
	// TODO: add test for logged in user. No user, no dice.
	//cart = models.Order{}

	DB(ctx).Preload("Product").Preload("Color").Preload("SizeVariations.Size").First(&colorVariation)
	//DB(ctx).Model(user).Related("Orders").First(&cart, "state = ?", "draft")
	//DB(ctx).Model(&cart).Related(user).First(&cart, "state = ?", "draft")
	err := DB(ctx).Where("user_id = ?", user.ID).First(cart, "state = ?", "checkout").Error
	if err != nil {
		log.Printf("No carts in database: %v", err)
	} else {
		err = DB(ctx).Where("order_id = ?", cart.ID).Find(&cart.OrderItems).Error
		if err != nil {
			log.Printf("No items in cart.")
		} else {
			for index, _ := range cart.OrderItems {
				err = DB(ctx).First(&cart.OrderItems[index].Product, cart.OrderItems[index].ProductID).Error
			}
		}

	}
	log.Printf("From cart controller--> cart contains: %v", cart)

	config.View.Funcs(funcsMap(ctx)).Execute(
		"cart_show",
		gin.H{
			"ActionBarTag":   admin.ActionBar.Render(ctx.Writer, ctx.Request),
			"Cart":           *cart,
			"ColorVariation": colorVariation,
			"CurrentUser":    CurrentUser(ctx),
			"CurrentLocale":  CurrentLocale(ctx),
		},
		ctx.Request,
		ctx.Writer,
	)
}

// func funcsMap(ctx *gin.Context) template.FuncMap {
