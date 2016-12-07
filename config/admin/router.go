package admin

import (
	"encoding/json"
	"fmt"

	"github.com/qor/admin"
	"github.com/sunwukonga/paypal-qor-admin/app/models"
)

type Charts struct {
	Tranx []models.Chart
	Users []models.Chart
}

func ReportsDataHandler(context *admin.Context) {
	charts := &Charts{}
	startDate := context.Request.URL.Query().Get("startDate")
	endDate := context.Request.URL.Query().Get("endDate")

	charts.Tranx = models.GetChartData("paypal_payments", startDate, endDate)
	charts.Users = models.GetChartData("users", startDate, endDate)

	b, _ := json.Marshal(charts)
	for _, v := range charts.Tranx {
		fmt.Println("Chart data: ", v.Total)
	}
	context.Writer.Write(b)
	return
}

func initRouter() {
	Admin.GetRouter().Get("/reports", ReportsDataHandler)
}
