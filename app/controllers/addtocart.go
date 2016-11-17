package controllers

import (
	//	"encoding/json"
	"log"
	"strconv"
	//	"html/template"
	//	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sunwukonga/qor-example/app/models"
	//	"github.com/sunwukonga/qor-example/config/auth" // for sessionStore
)

func AddToCart(ctx *gin.Context) {

	var (
		cart *models.Order
		user *models.User
		//sessionStorer *auth.SessionStorer
		productId uint
		product   models.Product
	)

	user = CurrentUser(ctx)
	cart = &models.Order{}

	productIdStr := ctx.Param("id")
	productIduint64, _ := strconv.ParseUint(productIdStr, 10, 32)
	productId = uint(productIduint64)

	// Does 'draft' Order for this user exist in the DB
	//err := DB(ctx).Model("Orders").Where().Find(cart, "state = ?", "draft").Error
	err := DB(ctx).Where("user_id = ?", user.ID).First(cart, "state = ?", "checkout").Error
	if err != nil {
		log.Printf("First record of Orders with state=checkout under the current user gave error: %v", err)
	} else {
		err = DB(ctx).Where("order_id = ?", cart.ID).Find(&cart.OrderItems).Error
		if err != nil {
			log.Printf("No items in cart.")
		}
	}

	if cart.UserID > 0 {
		// We have a 'checkout' Order that can be used as the cart.
		// Is the product in there already?
		// If product already exists in cart, ignore, else add to cart.
		// Quantity is never incremented.
		for _, orderItem := range cart.OrderItems {
			log.Println("Checking session cart for product ...")
			if orderItem.ProductID == productId {
				// signal that product should not be added to cart as it already exists
				productId = 0
				log.Println("We found the product in the cart already.")
				break
			}
		}
	} else {
		// Create a new draft order, i.e. cart
		cart = models.NewOrder(user, DB(ctx))
		log.Printf("Current state of cart after user added: %v", cart.GetState())
		DB(ctx).Create(cart)

	}

	if productId > 0 {
		// Fetch product details from the database.
		DB(ctx).First(&product, productId)
		log.Printf("Product %v fetched from database", productId)
		// add a new product to the cart
		cart.OrderItems = append(cart.OrderItems, *models.NewOrderItem(product, cart.ID, DB(ctx)))
		log.Println(cart)

		DB(ctx).Save(&cart)

	}

	redirectBack(ctx.Writer, ctx.Request)

}
