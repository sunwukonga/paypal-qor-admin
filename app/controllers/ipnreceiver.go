package controllers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func IpnReceiver(ctx *gin.Context) {
	//func IpnReceiver(w http.ResponseWriter, r *http.Request) {

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

}
