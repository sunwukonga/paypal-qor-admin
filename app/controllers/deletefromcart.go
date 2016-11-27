package controllers

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sunwukonga/paypal-qor-admin/app/models"
)

func DeleteFromCart(ctx *gin.Context) {

	// TODO: Check that currentUser owns this OrderItem.
	log.Println("We visited DeleteFromCart")
	orderItemIduint64, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)

	DB(ctx).Unscoped().Where("id = ?", uint(orderItemIduint64)).Delete(&models.OrderItem{})

	redirectBack(ctx.Writer, ctx.Request)
}
