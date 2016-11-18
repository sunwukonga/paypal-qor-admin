package controllers

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sunwukonga/qor-example/app/models"
)

func UpdateItemQty(ctx *gin.Context) {

	log.Printf("[UpdateItemQty] quantity: %v", ctx.Query("quantity"))
	// TODO: Check that currentUser owns this OrderItem.
	// TODO: Check that the state is checkout.
	orderItemIduint64, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	orderItemQtyint64, _ := strconv.ParseUint(ctx.PostForm("quantity"), 10, 32)

	// TODO: Used Model() instead of Table() as it offers more protection with respect to database changes into the future.
	DB(ctx).Model(&models.OrderItem{}).Where("id = ?", uint(orderItemIduint64)).Updates(map[string]interface{}{"quantity": uint(orderItemQtyint64)})

	redirectBack(ctx.Writer, ctx.Request)
}
