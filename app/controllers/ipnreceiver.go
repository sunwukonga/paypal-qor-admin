package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sunwukonga/paypal-qor-admin/app/models"
)

func IpnReceiver(ctx *gin.Context) {
	//func IpnReceiver(w http.ResponseWriter, r *http.Request) {
	var (
		paypalPayment *models.PaypalPayment
		paypalPayer   *models.PaypalPayer
		subscription  *models.Subscription
	)

	// Switch for production and live
	isProduction := false

	urlSimulator := "https://www.sandbox.paypal.com/cgi-bin/webscr"
	urlLive := "https://www.paypal.com/cgi-bin/webscr"
	paypalURL := urlSimulator

	if isProduction {
		paypalURL = urlLive
	}

	// *********************************************************
	// HANDSHAKE STEP 1 -- Write back an empty HTTP 200 response
	// *********************************************************
	fmt.Printf("Write Status 200")
	ctx.Writer.WriteHeader(http.StatusOK)

	// *********************************************************
	// HANDSHAKE STEP 2 -- Send POST data (IPN message) back as verification
	// *********************************************************
	// Get Content-Type of request to be parroted back to paypal
	contentType := ctx.Request.Header.Get("Content-Type")
	// Read the raw POST body
	body, _ := ioutil.ReadAll(ctx.Request.Body)
	// Prepend POST body with required field
	body = append([]byte("cmd=_notify-validate&"), body...)
	// Make POST request to paypal
	resp, _ := http.Post(paypalURL, contentType, bytes.NewBuffer(body))

	// *********************************************************
	// HANDSHAKE STEP 3 -- Read response for VERIFIED or INVALID
	// *********************************************************
	verifyStatus, _ := ioutil.ReadAll(resp.Body)

	// *********************************************************
	// Test for VERIFIED
	// *********************************************************
	if string(verifyStatus) != "VERIFIED" {
		log.Printf("Response: %v", string(verifyStatus))
		log.Println("This indicates that an attempt was made to spoof this interface, or we have a bug.")
		return
	}
	// We can now assume that the POSTed information in `body` is VERIFIED to be from Paypal.
	log.Printf("Response: %v", string(verifyStatus))

	values, _ := url.ParseQuery(string(body))
	for i, v := range values {
		fmt.Println(i, v)
	}

	// Grab custom data
	custom := map[string]string{}
	if err := json.Unmarshal([]byte(values["custom"][0]), &custom); err != nil {
		fmt.Println("Error, could not unmarshal JSON:", err.Error())
	}

	// Prepare copy of appropriate Subscription for use in Switch below.
	subscription = &models.Subscription{}
	if len(values["subscr_id"]) > 0 {
		if err := DB(ctx).Where("subscr_id = ?", values["subscr_id"][0]).First(subscription).Error; err != nil {
			// Not found or something more serious...
		}
	}

	if len(values["txn_type"]) > 0 {
		switch values["txn_type"][0] {
		case "subscr_signup":
			// Setup new Subscription
			if len(values["payer_id"]) > 0 {
				// Create or Fetch paypalPayer
				paypalPayer = models.NewPaypalPayer(values["payer_id"][0], DB(ctx))
				if paypalPayer.UserID == 0 {
					// We must create a new User
					paypalPayer.User.Email = values["payer_email"][0]
					paypalPayer.User.Name = sql.NullString{String: values["first_name"][0] + values["last_name"][0], Valid: true}
					paypalPayer.User.Role = models.RoleSubscriber
					paypalPayer.User.Confirmed = true
					paypalPayer.User.Addresses = []models.Address{
						models.Address{
							ContactName: paypalPayer.User.Name.String,
							Country:     values["address_country"][0],
							City:        values["address_city"][0],
							Address1:    values["address_street"][0],
							Postcode:    values["address_zip"][0],
						},
					}
					DB(ctx).Create(paypalPayer)
				}
				// Create new Subscription
				subscription = models.NewSubscription(custom["coupon"], paypalPayer, DB(ctx))
				subscription.SubscrID = values["subscr_id"][0]
				recurTimes, _ := strconv.Atoi(values["recur_times"][0])
				subscription.RecurTimes = recurTimes
				subscription.Period = values["period3"][0]
				subscription.SubscrDate = values["subscr_date"][0]
				DB(ctx).Create(subscription)
			}

		case "subscr_payment":
			// Create PaypalPayment if the payment has been successful.
			if len(values["payment_status"]) > 0 {
				if values["payment_status"][0] == "Completed" {
					//Success
					if len(values["payer_id"]) > 0 {
						// Create or Fetch paypalPayer
						paypalPayer = models.NewPaypalPayer(values["payer_id"][0], DB(ctx))
						if paypalPayer.UserID == 0 {
							// We must create a new User
							paypalPayer.User.Email = values["payer_email"][0]
							paypalPayer.User.Name = sql.NullString{String: values["first_name"][0] + values["last_name"][0], Valid: true}
							paypalPayer.User.Role = models.RoleSubscriber
							paypalPayer.User.Confirmed = true
							paypalPayer.User.Addresses = []models.Address{
								models.Address{
									ContactName: paypalPayer.User.Name.String,
									Country:     values["address_country"][0],
									City:        values["address_city"][0],
									Address1:    values["address_street"][0],
									Postcode:    values["address_zip"][0],
								},
							}
							DB(ctx).Create(paypalPayer)
						}
						// Create new PaypalPayment
						paypalPayment = models.NewPaypalPayment(custom["coupon"], paypalPayer, DB(ctx))
						paypalPayment.TxnID = values["txn_id"][0]
						paypalPayment.SubscrID = values["subscr_id"][0]
						gross, _ := strconv.ParseFloat(values["mc_gross"][0], 32)
						paypalPayment.McGross = float32(gross)
						fee, _ := strconv.ParseFloat(values["mc_fee"][0], 32)
						paypalPayment.McFee = float32(fee)
						paypalPayment.PaymentStatus = values["payment_status"][0]
						DB(ctx).Create(paypalPayment)
						if len(subscription.SubscrID) > 0 {
							// Subscription already exists, our state machine is in known state. We can call Trigger.
							models.SubscriptionState.Trigger("payment", subscription, DB(ctx))
						}

						// TODO: <Long term> Create a new Order to accomodate this payment and fastforward to "paid"
						// Belay that. This will be a future feature. Temporarily dropping the Delivery address, name, and phone
						// information in favour of a paypal note for now.
					} else {
						//Error: We have nothing to identify the payment with.
					}
				} else {
					// Log any payment_status that mean we haven't been paid.
				}
				// Ignore this as it means paypal sent con-conforming data.
			}
		case "subscr_cancel":
			// Record a cancel event against the Subscription. Wait for subscr_eot...
			//Get the referred to Subscription.
			if len(subscription.SubscrID) > 0 {
				models.SubscriptionState.Trigger("cancel", subscription, DB(ctx))
			} else {
				// Log attempt to cancel a subscription without reference to "subscr_id"
			}
		case "subscr_failed":
			// Enter "unpaid" state
			if len(subscription.SubscrID) > 0 {
				models.SubscriptionState.Trigger("fail", subscription, DB(ctx))
			} else {
				// Log attempt to cancel a subscription without reference to "subscr_id"
			}
		case "subscr_eot":
			// Enter "eot" (end of term) state
			if len(subscription.SubscrID) > 0 {
				models.SubscriptionState.Trigger("eot", subscription, DB(ctx))
			} else {
				// Log attempt to end a subscription without reference to "subscr_id"
			}
		case "subscr_modify":
			// We are not expecting this. Log the occurance.
		} // End switch
	} // End test for txn_type
} // End func
