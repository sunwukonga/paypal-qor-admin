package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/transition"
	"github.com/sunwukonga/paypal-qor-admin/config"
	//	"github.com/sunwukonga/paypal-qor-admin/app/models"
)

type Subscription struct {
	gorm.Model
	UserID       uint
	User         User `gorm:"save_associations:false"`
	InfluencerID uint
	Influencer   User `gorm:"save_associations:false"`

	//SubscrPayments []PaypalPayment `gorm:"foreignkey:SubscrID;AssociationForeignKey:SubscrID"`
	SubscrPayments []PaypalPayment `gorm:"ForeignKey:SubscrID"`

	SubscrID   string
	RecurTimes int
	Period     string
	SubscrDate time.Time

	CancelledAt time.Time
	EotAt       time.Time
	transition.Transition
}

type PaypalPayment struct {
	gorm.Model
	UserID       uint
	User         User `gorm:"save_associations:false"`
	InfluencerID uint
	Influencer   User `gorm:"save_associations:false"`

	TxnID         string
	SubscrID      string
	McCurrency    string
	McGross       float32
	McFee         float32
	PaymentStatus string
}

type PaypalPayer struct {
	gorm.Model
	PayerID string `sql:"size:13" gorm:"not null;unique;"`
	UserID  uint
	User    User
}

func (payment PaypalPayment) Net() float32 {
	return payment.McGross - payment.McFee
}

// NewPaypalPayer Create new PaypalPayer with PayerID, or fetch existing PaypalPayer
func NewPaypalPayer(payer string, db *gorm.DB) *PaypalPayer {
	// Check if PayerID exists in db "paypal_payment"
	paypalPayer := &PaypalPayer{}
	if err := db.Where("payer_id = ?", payer).First(paypalPayer).Error; err != nil {
		if err.Error() == "record not found" {
			// Populate a new one
			paypalPayer.PayerID = payer
			//		paypalPayer.User = User{}
		} else {
			// We have a real error...
			fmt.Printf("While creating New PaypalPayer: %v\n", err.Error())
			return nil
		}
	} else {
		//We found a PaypalPayer with this PayerID, he or she already exists in our database.
	}
	return paypalPayer

}

// NewSubscription Creat Subscription based on coupon and PaypalPayer
func NewSubscription(coupon string, paypalPayer *PaypalPayer, db *gorm.DB) *Subscription {
	var influencerCoupon InfluencerCoupon
	subscription := Subscription{}
	if len(coupon) == 6 {
		if err := db.First(&influencerCoupon, "code = ?", coupon).Error; err != nil {
			if err.Error() == "record not found" {
				// We have a problem. The coupon was not found, which should be impossible.
				fmt.Printf("While creating New Subscription, coupon not found: %v\n", err.Error())
			} else {
				// We have a real error...
				fmt.Printf("While creating New Subscription: %v\n", err.Error())
			}
		} else {
			// found it! Set InfluencerID
			subscription.InfluencerID = influencerCoupon.UserID
		}
	} else {
		// Not a valid coupon, leave the InfluencerID empty.
	}

	subscription.UserID = paypalPayer.UserID
	SubscriptionState.Trigger("signup", &subscription, db)
	return &subscription
}

// NewPaypalPayment Creat PaypalPayment based on coupon and PaypalPayer
func NewPaypalPayment(coupon string, paypalPayer *PaypalPayer, db *gorm.DB) *PaypalPayment {
	var influencerCoupon InfluencerCoupon
	paypalPayment := PaypalPayment{}
	if len(coupon) == 6 {
		if err := db.First(&influencerCoupon, "code = ?", coupon).Error; err != nil {
			if err.Error() == "record not found" {
				// We have a problem. The coupon was not found, which should be impossible.
				fmt.Printf("While creating New Subscription, coupon not found: %v\n", err.Error())
			} else {
				// We have a real error...
				fmt.Printf("While creating New Subscription: %v\n", err.Error())
			}
		} else {
			// found it! Set InfluencerID
			paypalPayment.InfluencerID = influencerCoupon.UserID
		}
	} else {
		// Not a valid coupon, leave the InfluencerID empty.
	}

	paypalPayment.UserID = paypalPayer.UserID
	return &paypalPayment
}

var (
	SubscriptionState = transition.New(&Subscription{})
)

func init() {
	// Define Subscription's States
	SubscriptionState.Initial("draft")
	SubscriptionState.State("active")
	SubscriptionState.State("cancelled").Enter(func(value interface{}, tx *gorm.DB) error {
		tx.Model(value).UpdateColumn("cancelled_at", time.Now().In(config.SGT))
		return nil
	})
	SubscriptionState.State("unpaid")
	SubscriptionState.State("eot")

	// Define Subscription's Events
	SubscriptionState.Event("signup").To("active").From("draft")

	paymentEvent := SubscriptionState.Event("payment")
	paymentEvent.To("active").From("active")
	paymentEvent.To("active").From("unpaid")

	SubscriptionState.Event("cancel").To("cancelled").From("active") // Don't include `From("unpaid")`, as I don't want to lose that info.
	SubscriptionState.Event("fail").To("unpaid").From("active", "unpaid")
	SubscriptionState.Event("eot").To("eot").From("active", "unpaid", "cancelled")

}
