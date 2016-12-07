package admin

import (
	"html/template"

	"github.com/qor/admin"
)

func initFuncMap() {
	Admin.RegisterFuncMap("render_latest_tranx", renderLatestTranx)
}

func renderLatestTranx(context *admin.Context) template.HTML {
	var tranxContext = context.NewResourceContext("PaypalPayment")
	tranxContext.Searcher.Pagination.PerPage = 5
	// orderContext.SetDB(orderContext.GetDB().Where("state in (?)", []string{"paid"}))

	if tranx, err := tranxContext.FindMany(); err == nil {
		return tranxContext.Render("index/table", tranx)
	}
	return template.HTML("")
}
