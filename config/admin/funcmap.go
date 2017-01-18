package admin

import (
	"html/template"

	"github.com/qor/admin"
	"github.com/sunwukonga/paypal-qor-admin/app/models"
)

func initFuncMap() {
	Admin.RegisterFuncMap("render_latest_tranx", renderLatestTranx)
	Admin.RegisterFuncMap("render_coupon_code", renderCouponCode)
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

func renderCouponCode(context *admin.Context) template.HTML {
	influencerCoupon := &models.InfluencerCoupon{}
	if err := context.GetDB().Where("user_id = ?", context.CurrentUser.(*models.User).ID).First(influencerCoupon).Error; err == nil {
		return template.HTML(influencerCoupon.Code)
	} else {
		return template.HTML("")
	}
}
