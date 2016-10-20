// +build enterprise

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"enterprise.getqor.com/microsite"
	"enterprise.getqor.com/promotion"
	"github.com/fatih/color"
	"github.com/qor/media_library"
	"github.com/sunwukonga/qor-example/config/admin"
	"github.com/sunwukonga/qor-example/db"
)

/* How to run this script
   $ go run --tags enterprise db/seeds/enterprise.go db/seeds/seeds.go
*/

func main() {
	Tables := []interface{}{
		&promotion.Discount{},
		&promotion.Rule{},
		&promotion.Benefit{},
		&promotion.Coupon{},
		&promotion.BenefitRecord{},
		&admin.QorMicroSite{},
		&microsite.QorMicroSitePackage{},
	}

	TruncateTables(Tables...)
	createPromotion()
	createMicroSite()
}

func createPromotion() {
	for _, enterpriseData := range Seeds.Enterprises {
		begins, _ := time.Parse("2006-01-02 15:04:05", enterpriseData.Begins)
		expires, _ := time.Parse("2006-01-02 15:04:05", enterpriseData.Expires)

		enterprise := promotion.Discount{}
		enterprise.Name = enterpriseData.Name
		enterprise.Begins = &begins
		enterprise.Expires = &expires
		enterprise.RequiresCoupon = enterpriseData.RequiresCoupon
		enterprise.Unique = enterpriseData.Unique

		if err := db.DB.Create(&enterprise).Error; err != nil {
			log.Fatalf("create enterprise (%v) failure, got err %v", enterprise, err)
		}

		for _, couponData := range enterpriseData.Coupons {
			coupon := promotion.Coupon{}
			coupon.DiscountID = enterprise.ID
			coupon.Code = couponData.Code
			if err := db.DB.Create(&coupon).Error; err != nil {
				log.Fatalf("create coupon (%v) failure, got err %v", coupon, err)
			}
		}

		for _, ruleData := range enterpriseData.Rules {
			rule := promotion.Rule{}
			rule.DiscountID = enterprise.ID
			rule.Kind = ruleData.Kind
			rule.Value.Scan(ruleData.Value)
			if err := db.DB.Create(&rule).Error; err != nil {
				log.Fatalf("create rule (%v) failure, got err %v", rule, err)
			}
		}

		for _, benefitData := range enterpriseData.Benefits {
			benefit := promotion.Benefit{}
			benefit.DiscountID = enterprise.ID
			benefit.Kind = benefitData.Kind
			benefit.Value.Scan(benefitData.Value)
			if err := db.DB.Create(&benefit).Error; err != nil {
				log.Fatalf("create benefit (%v) failure, got err %v", benefit, err)
			}
		}
	}
}

func createMicroSite() {
	site := admin.QorMicroSite{microsite.QorMicroSite{}}
	site.Name.Scan("Campaign")
	site.URL.Scan("/:locale/campaign")
	var packages []microsite.QorMicroSitePackage
	pakDatas := []struct {
		Template string
		Time     string
	}{
		{Template: "/db/seeds/data/campaign.zip", Time: "2016-09-10 10:00:00"},
		{Template: "/db/seeds/data/campaign_start.zip", Time: "2016-09-20 10:00:00"},
		{Template: "/db/seeds/data/campaign_finish.zip", Time: "2016-09-25 10:00:00"},
	}

	for _, pakData := range pakDatas {
		pak := microsite.QorMicroSitePackage{Template: media_library.FileSystem{}}
		file, err := os.Open(Root + pakData.Template)
		if err != nil {
			fmt.Printf(color.RedString(fmt.Sprintf("\nAccess MicroSite: can't open zip file, got (%s)\n", err)))
		}
		pak.Template.Scan(file)
		if pakData.Time != "" {
			t, _ := time.Parse("2006-01-02 15:04:05", pakData.Time)
			pak.StartAt = &t
		}
		packages = append(packages, pak)
	}
	site.Packages = packages
	widgets := []microsite.WidgetSetting{}
	for _, widgetName := range []string{"TopBanner", "FeatureProducts"} {
		widgets = append(widgets, microsite.WidgetSetting{Name: widgetName})
	}
	site.Widgets = microsite.WidgetBox{Widgets: widgets}
	if err := db.DB.Create(&site).Error; err != nil {
		log.Fatalf("create microsite (%v) failure, got err %v", site, err)
	}
}
