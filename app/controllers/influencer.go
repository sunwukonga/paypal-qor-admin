package controllers

import (
	"io"
	"net/http"

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

	if origin := ctx.Request.Header.Get("Origin"); origin != "" {
		ctx.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token")
	}
	ctx.Writer.WriteHeader(http.StatusOK)
	io.WriteString(ctx.Writer, exists)
}
