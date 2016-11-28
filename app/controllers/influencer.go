package controllers

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/sunwukonga/paypal-qor-admin/app/models"
)

func CodeExists(ctx *gin.Context) {
	var (
		user   models.User
		code   string
		exists string
	)

	code = string(ctx.Param("code"))

	DB(ctx).Where(&models.User{InfluencerCode: code}).First(&user)
	if user.InfluencerCode == code {
		exists = "true"
	} else {
		exists = "false"
	}

	io.WriteString(ctx.Writer, exists)
}
