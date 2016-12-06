package admin

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jinzhu/gorm"
	"github.com/qor/action_bar"
	"github.com/qor/activity"
	"github.com/qor/admin"
	"github.com/qor/i18n/exchange_actions"
	"github.com/qor/l10n/publish"
	"github.com/qor/media_library"
	"github.com/qor/notification"
	"github.com/qor/notification/channels/database"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
	"github.com/qor/transition"
	"github.com/qor/validations"
	"github.com/sunwukonga/paypal-qor-admin/app/models"
	"github.com/sunwukonga/paypal-qor-admin/config/admin/bindatafs"
	"github.com/sunwukonga/paypal-qor-admin/config/auth"
	"github.com/sunwukonga/paypal-qor-admin/config/i18n"
	"github.com/sunwukonga/paypal-qor-admin/db"
)

var Admin *admin.Admin
var ActionBar *action_bar.ActionBar
var Countries = []string{"China", "Japan", "USA"}
var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func init() {
	Admin = admin.New(&qor.Config{DB: db.DB.Set("publish:draft_mode", true)})
	Admin.SetSiteName("Qor DEMO")
	Admin.SetAuth(auth.AdminAuth{})
	Admin.SetAssetFS(bindatafs.AssetFS)

	// Add Notification
	Notification := notification.New(&notification.Config{})
	Notification.RegisterChannel(database.New(&database.Config{DB: db.DB}))
	Notification.Action(&notification.Action{
		Name: "Confirm",
		Visible: func(data *notification.QorNotification, context *admin.Context) bool {
			return data.ResolvedAt == nil
		},
		MessageTypes: []string{"order_returned"},
		Handle: func(argument *notification.ActionArgument) error {
			orderID := regexp.MustCompile(`#(\d+)`).FindStringSubmatch(argument.Message.Body)[1]
			err := argument.Context.GetDB().Model(&models.Order{}).Where("id = ? AND returned_at IS NULL", orderID).Update("returned_at", time.Now()).Error
			if err == nil {
				return argument.Context.GetDB().Model(argument.Message).Update("resolved_at", time.Now()).Error
			}
			return err
		},
		Undo: func(argument *notification.ActionArgument) error {
			orderID := regexp.MustCompile(`#(\d+)`).FindStringSubmatch(argument.Message.Body)[1]
			err := argument.Context.GetDB().Model(&models.Order{}).Where("id = ? AND returned_at IS NOT NULL", orderID).Update("returned_at", nil).Error
			if err == nil {
				return argument.Context.GetDB().Model(argument.Message).Update("resolved_at", nil).Error
			}
			return err
		},
	})
	Notification.Action(&notification.Action{
		Name:         "Check it out",
		MessageTypes: []string{"order_paid_cancelled", "order_processed", "order_returned"},
		URL: func(data *notification.QorNotification, context *admin.Context) string {
			return path.Join("/admin/orders/", regexp.MustCompile(`#(\d+)`).FindStringSubmatch(data.Body)[1])
		},
	})
	Notification.Action(&notification.Action{
		Name:         "Dismiss",
		MessageTypes: []string{"order_paid_cancelled", "info", "order_processed", "order_returned"},
		Visible: func(data *notification.QorNotification, context *admin.Context) bool {
			return data.ResolvedAt == nil
		},
		Handle: func(argument *notification.ActionArgument) error {
			return argument.Context.GetDB().Model(argument.Message).Update("resolved_at", time.Now()).Error
		},
		Undo: func(argument *notification.ActionArgument) error {
			return argument.Context.GetDB().Model(argument.Message).Update("resolved_at", nil).Error
		},
	})
	Admin.NewResource(Notification)

	// Add Dashboard
	Admin.AddMenu(&admin.Menu{Name: "Dashboard", Link: "/admin"})

	// Add Asset Manager, for rich editor
	assetManager := Admin.AddResource(&media_library.AssetManager{}, &admin.Config{Invisible: true})

	//* Produc Management *//
	Admin.AddMenu(&admin.Menu{Name: "Product Management", Link: "/productmanagement", Ancestors: []string{"Dashboard"}, Permission: roles.Deny(roles.CRUD, "servicer")})
	color := Admin.AddResource(&models.Color{}, &admin.Config{Menu: []string{"Product Management"}, Permission: roles.Deny(roles.CRUD, "servicer"), Priority: -5})
	Admin.AddResource(&models.Size{}, &admin.Config{Menu: []string{"Product Management"}, Permission: roles.Deny(roles.CRUD, "servicer"), Priority: -4})

	category := Admin.AddResource(&models.Category{}, &admin.Config{Menu: []string{"Product Management"}, Permission: roles.Deny(roles.CRUD, "servicer"), Priority: -3})
	category.Meta(&admin.Meta{Name: "Categories", Type: "select_many"})

	collection := Admin.AddResource(&models.Collection{}, &admin.Config{Menu: []string{"Product Management"}, Permission: roles.Deny(roles.CRUD, "servicer"), Priority: -2})

	// Add ProductImage as Media Libraray
	ProductImagesResource := Admin.AddResource(&models.ProductImage{}, &admin.Config{Menu: []string{"Product Management"}, Permission: roles.Deny(roles.CRUD, "servicer"), Priority: -1})

	ProductImagesResource.Filter(&admin.Filter{
		Name:   "SelectedType",
		Label:  "Media Type",
		Config: &admin.SelectOneConfig{Collection: [][]string{{"video", "Video"}, {"image", "Image"}, {"file", "File"}, {"video_link", "Video Link"}}},
	})
	ProductImagesResource.Filter(&admin.Filter{
		Name:   "Color",
		Config: &admin.SelectOneConfig{RemoteDataResource: color},
	})
	ProductImagesResource.Filter(&admin.Filter{
		Name:   "Category",
		Config: &admin.SelectOneConfig{RemoteDataResource: category},
	})
	ProductImagesResource.IndexAttrs("File", "Title")

	// Add Product
	product := Admin.AddResource(&models.Product{}, &admin.Config{Menu: []string{"Product Management"}, Permission: roles.Deny(roles.CRUD, "servicer")})
	product.Meta(&admin.Meta{Name: "MadeCountry", Config: &admin.SelectOneConfig{Collection: Countries}})
	product.Meta(&admin.Meta{Name: "Description", Config: &admin.RichEditorConfig{AssetManager: assetManager, Plugins: []admin.RedactorPlugin{
		{Name: "medialibrary", Source: "/admin/assets/javascripts/qor_redactor_medialibrary.js"},
		{Name: "table", Source: "/javascripts/redactor_table.js"},
	},
		Settings: map[string]interface{}{
			"medialibraryUrl": "/admin/product_images",
		},
	}})
	product.Meta(&admin.Meta{Name: "Category", Config: &admin.SelectOneConfig{AllowBlank: true}})
	product.Meta(&admin.Meta{Name: "Collections", Config: &admin.SelectManyConfig{SelectMode: "bottom_sheet"}})

	product.Meta(&admin.Meta{Name: "MainImage", Config: &media_library.MediaBoxConfig{
		RemoteDataResource: ProductImagesResource,
		Max:                1,
		Sizes: map[string]media_library.Size{
			"preview": {Width: 300, Height: 300},
		},
	}})
	product.Meta(&admin.Meta{Name: "MainImageURL", Valuer: func(record interface{}, context *qor.Context) interface{} {
		if p, ok := record.(*models.Product); ok {
			result := bytes.NewBufferString("")
			tmpl, _ := template.New("").Parse("<img src='{{.image}}'></img>")
			tmpl.Execute(result, map[string]string{"image": p.MainImageURL()})
			return template.HTML(result.String())
		}
		return ""
	}})

	product.Filter(&admin.Filter{
		Name:   "Collections",
		Config: &admin.SelectOneConfig{RemoteDataResource: collection},
	})

	product.UseTheme("grid")

	colorVariationMeta := product.Meta(&admin.Meta{Name: "ColorVariations"})
	colorVariation := colorVariationMeta.Resource
	colorVariation.Meta(&admin.Meta{Name: "Images", Config: &media_library.MediaBoxConfig{
		RemoteDataResource: ProductImagesResource,
		Sizes: map[string]media_library.Size{
			"icon":    {Width: 50, Height: 50},
			"preview": {Width: 300, Height: 300},
			"listing": {Width: 640, Height: 640},
		},
	}})

	colorVariation.NewAttrs("-Product", "-ColorCode")
	colorVariation.EditAttrs("-Product", "-ColorCode")

	sizeVariationMeta := colorVariation.Meta(&admin.Meta{Name: "SizeVariations"})
	sizeVariation := sizeVariationMeta.Resource
	sizeVariation.NewAttrs("-ColorVariation")
	sizeVariation.EditAttrs(
		&admin.Section{
			Rows: [][]string{
				{"Size", "AvailableQuantity"},
			},
		},
	)

	product.SearchAttrs("Name", "Code", "Category.Name", "Brand.Name")
	product.IndexAttrs("MainImageURL", "Name", "Price")
	product.EditAttrs(
		&admin.Section{
			Title: "Basic Information",
			Rows: [][]string{
				{"Name"},
				{"Code", "Price"},
				{"MainImage"},
			}},
		&admin.Section{
			Title: "Organization",
			Rows: [][]string{
				{"Category", "MadeCountry"},
				{"Collections"},
			}},
		"ProductProperties",
		"Description",
		"ColorVariations",
	)
	product.NewAttrs(product.EditAttrs())

	for _, country := range Countries {
		var country = country
		product.Scope(&admin.Scope{Name: country, Group: "Made Country", Handle: func(db *gorm.DB, ctx *qor.Context) *gorm.DB {
			return db.Where("made_country = ?", country)
		}})
	}

	product.Action(&admin.Action{
		Name: "View On Site",
		URL: func(record interface{}, context *admin.Context) string {
			if product, ok := record.(*models.Product); ok {
				return fmt.Sprintf("/products/%v", product.Code)
			}
			return "#"
		},
		Modes: []string{"menu_item", "edit"},
	})

	product.Action(&admin.Action{
		Name: "Disable",
		Handle: func(arg *admin.ActionArgument) error {
			for _, record := range arg.FindSelectedRecords() {
				arg.Context.DB.Model(record.(*models.Product)).Update("enabled", false)
			}
			return nil
		},
		Visible: func(record interface{}, context *admin.Context) bool {
			if product, ok := record.(*models.Product); ok {
				return product.Enabled == true
			}
			return true
		},
		Modes: []string{"index", "edit", "menu_item"},
	})

	product.Action(&admin.Action{
		Name: "Enable",
		Handle: func(arg *admin.ActionArgument) error {
			for _, record := range arg.FindSelectedRecords() {
				arg.Context.DB.Model(record.(*models.Product)).Update("enabled", true)
			}
			return nil
		},
		Visible: func(record interface{}, context *admin.Context) bool {
			if product, ok := record.(*models.Product); ok {
				return product.Enabled == false
			}
			return true
		},
		Modes: []string{"index", "edit", "menu_item"},
	})

	// Add Order
	order := Admin.AddResource(&models.Order{}, &admin.Config{Menu: []string{"Order Management"}})
	order.Meta(&admin.Meta{Name: "ShippingAddress", Type: "single_edit"})
	order.Meta(&admin.Meta{Name: "BillingAddress", Type: "single_edit"})
	order.Meta(&admin.Meta{Name: "ShippedAt", Type: "date"})

	orderItemMeta := order.Meta(&admin.Meta{Name: "OrderItems"})
	orderItemMeta.Resource.Meta(&admin.Meta{Name: "SizeVariation", Config: &admin.SelectOneConfig{Collection: sizeVariationCollection}})

	// define scopes for Order
	for _, state := range []string{"checkout", "cancelled", "paid", "paid_cancelled", "processing", "shipped", "returned"} {
		var state = state
		order.Scope(&admin.Scope{
			Name:  state,
			Label: strings.Title(strings.Replace(state, "_", " ", -1)),
			Group: "Order Status",
			Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
				return db.Where(models.Order{Transition: transition.Transition{State: state}})
			},
		})
	}

	// define actions for Order
	type trackingNumberArgument struct {
		TrackingNumber string
	}

	order.Action(&admin.Action{
		Name: "Processing",
		Handle: func(argument *admin.ActionArgument) error {
			for _, order := range argument.FindSelectedRecords() {
				db := argument.Context.GetDB()
				if err := models.OrderState.Trigger("process", order.(*models.Order), db); err != nil {
					return err
				}
				db.Select("state").Save(order)
			}
			return nil
		},
		Visible: func(record interface{}, context *admin.Context) bool {
			if order, ok := record.(*models.Order); ok {
				return order.State == "paid"
			}
			return false
		},
		Modes: []string{"show", "menu_item"},
	})
	order.Action(&admin.Action{
		Name: "Ship",
		Handle: func(argument *admin.ActionArgument) error {
			var (
				tx                     = argument.Context.GetDB().Begin()
				trackingNumberArgument = argument.Argument.(*trackingNumberArgument)
			)

			if trackingNumberArgument.TrackingNumber != "" {
				for _, record := range argument.FindSelectedRecords() {
					order := record.(*models.Order)
					order.TrackingNumber = &trackingNumberArgument.TrackingNumber
					models.OrderState.Trigger("ship", order, tx, "tracking number "+trackingNumberArgument.TrackingNumber)
					if err := tx.Save(order).Error; err != nil {
						tx.Rollback()
						return err
					}
				}
			} else {
				return errors.New("invalid shipment number")
			}

			tx.Commit()
			return nil
		},
		Visible: func(record interface{}, context *admin.Context) bool {
			if order, ok := record.(*models.Order); ok {
				return order.State == "processing"
			}
			return false
		},
		Resource: Admin.NewResource(&trackingNumberArgument{}),
		Modes:    []string{"show", "menu_item"},
	})

	order.Action(&admin.Action{
		Name: "Cancel",
		Handle: func(argument *admin.ActionArgument) error {
			for _, order := range argument.FindSelectedRecords() {
				db := argument.Context.GetDB()
				if err := models.OrderState.Trigger("cancel", order.(*models.Order), db); err != nil {
					return err
				}
				db.Select("state").Save(order)
			}
			return nil
		},
		Visible: func(record interface{}, context *admin.Context) bool {
			if order, ok := record.(*models.Order); ok {
				for _, state := range []string{"draft", "checkout", "paid", "processing"} {
					if order.State == state {
						return true
					}
				}
			}
			return false
		},
		Modes: []string{"index", "show", "menu_item"},
	})

	order.IndexAttrs("User", "PaymentAmount", "ShippedAt", "CancelledAt", "State", "ShippingAddress")
	order.NewAttrs("-DiscountValue", "-AbandonedReason", "-CancelledAt")
	order.EditAttrs("-DiscountValue", "-AbandonedReason", "-CancelledAt", "-State")
	order.ShowAttrs("-DiscountValue", "-State")
	order.SearchAttrs("User.Name", "User.Email", "ShippingAddress.ContactName", "ShippingAddress.Address1", "ShippingAddress.Address2")

	// Add activity for order
	activity.Register(order)

	// Define another resource for same model
	abandonedOrder := Admin.AddResource(&models.Order{}, &admin.Config{Name: "Abandoned Order", Menu: []string{"Order Management"}})
	abandonedOrder.Meta(&admin.Meta{Name: "ShippingAddress", Type: "single_edit"})
	abandonedOrder.Meta(&admin.Meta{Name: "BillingAddress", Type: "single_edit"})

	// Define default scope for abandoned orders
	abandonedOrder.Scope(&admin.Scope{
		Default: true,
		Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("abandoned_reason IS NOT NULL AND abandoned_reason <> ?", "")
		},
	})

	// Define scopes for abandoned orders
	for _, amount := range []int{5000, 10000, 20000} {
		var amount = amount
		abandonedOrder.Scope(&admin.Scope{
			Name:  fmt.Sprint(amount),
			Group: "Amount Greater Than",
			Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
				return db.Where("payment_amount > ?", amount)
			},
		})
	}

	abandonedOrder.IndexAttrs("-ShippingAddress", "-BillingAddress", "-DiscountValue", "-OrderItems")
	abandonedOrder.NewAttrs("-DiscountValue")
	abandonedOrder.EditAttrs("-DiscountValue")
	abandonedOrder.ShowAttrs("-DiscountValue")

	// Add Beauty Box Transactions
	payments := Admin.AddResource(&models.PaypalPayment{}, &admin.Config{Name: "Transactions", Menu: []string{"Beauty Box"}})
	payments.Meta(&admin.Meta{
		Name:  "SubscrID",
		Label: "Subscription",
		Type:  "readonly",
	})
	payments.Meta(&admin.Meta{
		Name:  "UserID",
		Label: "Subscriber",
		Type:  "readonly",
	})
	payments.Meta(&admin.Meta{
		Name:  "InfluencerID",
		Label: "Influencer",
		Type:  "readonly",
	})
	payments.Meta(&admin.Meta{
		Name:  "PaymentStatus",
		Label: "Status",
		Type:  "readonly",
	})

	payments.IndexAttrs("SubscrID", "UserID", "InfluencerID", "PaymentStatus")
	payments.NewAttrs()
	payments.EditAttrs()
	payments.ShowAttrs(payments.IndexAttrs())

	// Add Beauty Box Subscriptions
	subscriptions := Admin.AddResource(&models.Subscription{}, &admin.Config{Name: "Subscriptions", Menu: []string{"Beauty Box"}})
	subscriptions.Meta(&admin.Meta{
		Name:  "SubscrID",
		Label: "ID",
		Type:  "readonly",
	})
	subscriptions.Meta(&admin.Meta{
		Name:  "UserID",
		Label: "Subscriber",
		Type:  "readonly",
	})
	subscriptions.Meta(&admin.Meta{
		Name:  "InfluencerID",
		Label: "Influencer",
		Type:  "readonly",
	})
	subscriptions.Meta(&admin.Meta{
		Name:  "Period",
		Label: "Period",
		Type:  "readonly",
	})
	subscriptions.Meta(&admin.Meta{
		Name:  "SubscrDate",
		Label: "Signup Date",
		Type:  "readonly",
	})
	subscriptions.Meta(&admin.Meta{
		Name:  "RecurTimes",
		Label: "Total",
		Type:  "readonly",
	})
	subscriptions.Meta(&admin.Meta{
		Name:  "State",
		Label: "Status",
		Type:  "readonly",
	})
	subscriptions.Meta(&admin.Meta{
		Name:  "CancelledAt",
		Label: "Cancel Date",
		Type:  "readonly",
	})
	subscriptions.Meta(&admin.Meta{
		Name:  "EotAt",
		Label: "End Date",
		Type:  "readonly",
	})
	associatedTransactions := subscriptions.Meta(&admin.Meta{
		Name:  "SubscrPayments",
		Label: "Transactions",
		Type:  "collection_edit",
		Valuer: func(record interface{}, context *qor.Context) interface{} {
			paypalPayments := &[]models.PaypalPayment{}
			subscription := record.(*models.Subscription)
			context.GetDB().Where("subscr_id = ?", subscription.SubscrID).Find(paypalPayments)
			return paypalPayments
		},
	}).Resource
	associatedTransactions.Meta(&admin.Meta{
		Name:  "NetPayment",
		Label: "Net Payment",
		Type:  "readonly",
		Valuer: func(record interface{}, context *qor.Context) interface{} {
			txn := record.(*models.PaypalPayment)
			return txn.Net()
		},
	})
	associatedTransactions.Meta(&admin.Meta{
		Name: "McFee",
		Type: "hidden",
	})
	associatedTransactions.Meta(&admin.Meta{
		Name: "McGross",
		Type: "hidden",
	})
	associatedTransactions.Meta(&admin.Meta{
		Name: "User",
		Type: "hidden",
	})
	associatedTransactions.Meta(&admin.Meta{
		Name: "Influencer",
		Type: "hidden",
	})
	associatedTransactions.ShowAttrs(
		&admin.Section{
			Title: "Tranx",
			Rows: [][]string{
				{"ID", "NetPayment", "Status"},
			},
		},
	)
	associatedTransactions.EditAttrs(associatedTransactions.ShowAttrs)
	associatedTransactions.IndexAttrs(associatedTransactions.ShowAttrs)

	/*	associatedTransactions.Meta(&admin.Meta{
			Name:  "TxnID",
			Label: "ID",
			Type:  "readonly",
		})
		associatedTransactions.Meta(&admin.Meta{
			Name: "Net Payment",
			Type: "float*",
			Valuer: func(record interface{}, context *qor.Context) interface{} {
				fmt.Println(record)
				//payment := record.(*models.PaypalPayment)
				return "" // payment.Net()
			},
		})
	*/
	subscriptions.IndexAttrs("SubscrID", "UserID", "InfluencerID", "State")
	subscriptions.NewAttrs()
	subscriptions.EditAttrs()
	subscriptions.ShowAttrs(
		&admin.Section{
			Title: "Subscription",
			Rows: [][]string{
				{"SubscrID", "State"},
				{"UserID", "InfluencerID"},
				{"RecurTimes", "Period"},
				{"SubscrDate"},
				{"CancelledAt", "EotAt"},
			}},
		"SubscrPayments",
	)
	// Add User
	user := Admin.AddResource(&models.User{}, &admin.Config{Menu: []string{"User Management"}})
	user.Action(&admin.Action{
		Name: "GenerateCode",
		Handle: func(argument *admin.ActionArgument) error {
			var (
				tx = argument.Context.GetDB().Begin()
			)

			for _, record := range argument.FindSelectedRecords() {
				user := record.(*models.User)
				// Create a new code.
				// Check that the code does not already exist. Very, very unlikely.
				// Insert code into database.
				influencerCoupon := &models.InfluencerCoupon{}
				if err := tx.Where("user_id = ?", user.ID).First(influencerCoupon).Error; err != nil {
					//Inspect error. Hopefully, that it wasn't found.
					// Create Coupon
					n := 6
					b := make([]byte, n)
					// http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
					for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
						if remain == 0 {
							cache, remain = src.Int63(), letterIdxMax
						}
						if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
							b[i] = letterBytes[idx]
							i--
						}
						cache >>= letterIdxBits
						remain--
					}
					influencerCoupon.Code = string(b)
					influencerCoupon.UserID = user.ID
					influencerCoupon.User = *user

					//Add InfluencerCoupon to DB
					if err := tx.Save(influencerCoupon).Error; err != nil {
						tx.Rollback()
						return err
					}
				} else {
					// User already has a coupon code. Not good. Nothing to do.
					fmt.Println("Error. We should not be here. Influencer already had coupon code.")
				}

			}

			tx.Commit()
			return nil
		},
		Visible: func(record interface{}, context *admin.Context) bool {
			if user, ok := record.(*models.User); ok {
				//return true if InfluencerCoupon doesn't exist, or it is invalid.
				if user.Role == models.RoleInfluencer {
					influencerCoupon := &models.InfluencerCoupon{}
					if err := context.GetDB().Where("user_id = ?", user.ID).First(influencerCoupon).Error; err != nil {
						if err.Error() == "record not found" {
							return true
						} else {
							// Ooops, we found a real error
							fmt.Println("Error fetching coupon: ", err.Error())
							return true
						}
					} else {
						// Following test for invalidity is not complete. Code should only contain [A-Z\d]
						if len(influencerCoupon.Code) != 6 {
							return true
						}
					}
				}
			}
			return false
		},
		Modes: []string{"show", "menu_item"},
	})
	user.Meta(&admin.Meta{Name: "Gender", Config: &admin.SelectOneConfig{Collection: []string{"Male", "Female", "Unknown"}}})
	// TODO: replace Role strings with string constants
	user.Meta(&admin.Meta{Name: "Role", Config: &admin.SelectOneConfig{Collection: models.Roles}})
	user.Meta(&admin.Meta{Name: "Password",
		Type:            "password",
		FormattedValuer: func(interface{}, *qor.Context) interface{} { return "" },
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			values := metaValue.Value.([]string)
			u := resource.(*models.User)
			if len(values) > 0 {
				if newPassword := values[0]; newPassword != "" {
					bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
					if err != nil {
						context.DB.AddError(validations.NewError(user, "Password", "Can't encrpt password"))
						return
					}
					u.Password = string(bcryptPassword)
				}
			}
			// Piggyback here to auto-fill the InfluencerCode field
			// Should only run if this code does not already exist.
			/*
				n := 6
				if len(u.InfluencerCode) != n {
					b := make([]byte, n)
					// http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
					for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
						if remain == 0 {
							cache, remain = src.Int63(), letterIdxMax
						}
						if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
							b[i] = letterBytes[idx]
							i--
						}
						cache >>= letterIdxBits
						remain--
					}
					u.InfluencerCode = string(b)
				}
			*/
		},
	})
	user.Meta(&admin.Meta{Name: "InfluencerCode",
		Label: "Influencer Code",
		Type:  "readonly",
		Valuer: func(record interface{}, context *qor.Context) interface{} {
			influencerCoupon := &models.InfluencerCoupon{}
			user := record.(*models.User)
			if err := context.GetDB().Where("user_id = ?", user.ID).First(influencerCoupon).Error; err != nil {
				if err.Error() == "record not found" {
					return ""
				} else {
					// Ooops, we found a real error
					fmt.Println("Error fetching coupon: ", err.Error())
					return ""
				}
			} else {
				return influencerCoupon.Code
			}
		},
	})
	user.Meta(&admin.Meta{Name: "Confirmed", Valuer: func(user interface{}, ctx *qor.Context) interface{} {
		if user.(*models.User).ID == 0 {
			return true
		}
		return user.(*models.User).Confirmed
	}})

	user.Filter(&admin.Filter{
		Name: "Role",
		Config: &admin.SelectOneConfig{
			Collection: models.Roles,
			//Collection: []string{"Admin", "Influencer", "Maintainer", "Member"},
		},
	})

	user.IndexAttrs("ID", "Email", "Name", "Gender", "Role")
	user.ShowAttrs(
		&admin.Section{
			Title: "Basic Information",
			Rows: [][]string{
				{"Name"},
				{"Email", "Password"},
				{"Gender", "Role"},
				{"InfluencerCode"},
				{"Confirmed"},
			}},
		"Addresses",
	)
	user.NewAttrs(
		&admin.Section{
			Title: "Basic Information",
			Rows: [][]string{
				{"Name"},
				{"Email", "Password"},
				{"Gender", "Role"},
				{"Confirmed"},
			}},
		"Addresses",
	)
	user.EditAttrs(user.NewAttrs())

	// Add Store
	store := Admin.AddResource(&models.Store{}, &admin.Config{Menu: []string{"Store Management"}})
	store.Meta(&admin.Meta{Name: "Owner", Type: "single_edit"})
	store.AddValidator(func(record interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
		if meta := metaValues.Get("Name"); meta != nil {
			if name := utils.ToString(meta.Value); strings.TrimSpace(name) == "" {
				return validations.NewError(record, "Name", "Name can't be blank")
			}
		}
		return nil
	})

	// Add Translations
	Admin.AddResource(i18n.I18n, &admin.Config{Menu: []string{"Site Management"}, Priority: 1})

	// Add SEOSetting
	Admin.AddResource(&models.SEOSetting{}, &admin.Config{Menu: []string{"Site Management"}, Singleton: true, Priority: 2})

	// Add Worker
	Worker := getWorker()
	Admin.AddResource(Worker, &admin.Config{Menu: []string{"Site Management"}})

	db.Publish.SetWorker(Worker)
	exchange_actions.RegisterExchangeJobs(i18n.I18n, Worker)

	// Add Publish
	Admin.AddResource(db.Publish, &admin.Config{Menu: []string{"Site Management"}, Singleton: true})
	publish.RegisterL10nForPublish(db.Publish, Admin)

	// Add Setting
	Admin.AddResource(&models.Setting{}, &admin.Config{Name: "Shop Setting", Singleton: true})

	// Add Search Center Resources
	Admin.AddSearchResource(product, user, order)

	// Add ActionBar
	ActionBar = action_bar.New(Admin, auth.AdminAuth{})
	ActionBar.RegisterAction(&action_bar.Action{Name: "Admin Dashboard", Link: "/admin"})

	initWidgets()
	initFuncMap()
	initRouter()
}

func sizeVariationCollection(resource interface{}, context *qor.Context) (results [][]string) {
	for _, sizeVariation := range models.SizeVariations() {
		results = append(results, []string{strconv.Itoa(int(sizeVariation.ID)), sizeVariation.Stringify()})
	}
	return
}
