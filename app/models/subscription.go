package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/transition"
	//	"github.com/sunwukonga/paypal-qor-admin/app/models"
)

type Subscription struct {
	gorm.Model
	UserID       uint
	User         User `gorm:"save_associations:false"`
	InfluencerID uint
	Influencer   User `gorm:"save_associations:false"`

	SubscrID   string
	RecurTimes int
	Period     string
	SubscrDate string

	CancelledAt *time.Time
	EotAt       *time.Time
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
	// Define Order's States
	SubscriptionState.Initial("draft")
	SubscriptionState.State("active")
	SubscriptionState.State("cancelled").Enter(func(value interface{}, tx *gorm.DB) error {
		tx.Model(value).UpdateColumn("cancelled_at", time.Now())
		return nil
	})
	SubscriptionState.State("unpaid").Enter(func(value interface{}, tx *gorm.DB) error {
		// This requires no action that I yet know of. We are waiting for the customer to resolve his or her payment difficulties.
		return nil
	})
	SubscriptionState.State("eot").Enter(func(value interface{}, tx *gorm.DB) error {
		// Perhaps we would require some clean up here?
		return nil
	})

	SubscriptionState.Event("signup").To("active").From("draft")

	paymentEvent := SubscriptionState.Event("payment")
	paymentEvent.To("active").From("active")
	paymentEvent.To("active").From("unpaid")

	SubscriptionState.Event("cancel").To("cancelled").From("active", "unpaid")
	SubscriptionState.Event("fail").To("unpaid").From("active", "unpaid")

	eotEvent := SubscriptionState.Event("eot")
	eotEvent.To("eot").From("active")
	eotEvent.To("eot").From("unpaid")
	eotEvent.To("eot").From("cancelled")

}
