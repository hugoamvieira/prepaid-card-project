package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hugoamvieira/prepaid-card-project/backend/src/prepaid"
)

const (
	// API Errors
	errCardNumberIsNaN          = `{"error": "Card number is not a Number"}`
	errCardNotFound             = `{"error": "We couldn't find a card with that number"}`
	errIncorrectOrMalformedJSON = `{"error": "Incorrect or malformed JSON."}`
	errRqInvalidOrMissingData   = `{"error": "Invalid or missing data. Try again."}`
	errCardCreation             = `{"error": "Sorry, we couldn't create your card. Try again."}`
	errInvalidCardNumProvided   = `{"error": "The card number provided is invalid."}`
	errBackendGeneric           = `{"error": "Internal server error."}`
	errNotEnoughFunds           = `{"error": "Not enough funds on card."}`
	errCardIsNotUsable          = `{"error": "This card has not been activated."}`
	errAuthRqNotFound           = `{"error": "We couldn't find an auth request with that UUID."}`
	errAuthReversed             = `{"error": "Authorization Request has been reversed."}`
	errInvalidCaptureAmount     = `{"error": "Invalid Capture amount."}`
	errInvalidReverseAmount     = `{"error": "Invalid Reverse amount."}`
	errFundCaptureNotFound      = `{"error": "Funds Capture not found."}`
	errInvalidRefundAmount      = `{"error": "Invalid Refund amount.}`

	// API Responses
	cardCreationSuccess = `{"card_number": "%v"}`
	authRequestSuccess  = `{"auth_request_uuid": "%v"}`
)

type createCardRequest struct {
	CardholderName string `json:"cardholder_name"`
	CardCurrency   string `json:"card_currency"`
}

type loadFundsRequest struct {
	AmountCents int64 `json:"amount_cents"`
}

type authRequest struct {
	MerchantName string `json:"merchant_name"`
	AuthAmount   int64  `json:"auth_amount"` // In Cents
}

type reverseAuthRequest struct {
	ReverseAmount int64 `json:"reverse_amount"`
}

type captureFundsRequest struct {
	CaptureAmount int64 `json:"capture_amount"`
}

type refundFundsRequest struct {
	RefundAmount int64 `json:"refund_amount"`
}

// NewAPI creates and starts the backend's API
func NewAPI() {
	r := mux.NewRouter()
	r.HandleFunc("/card/new", createCardAPIHandler).Methods("POST")
	r.HandleFunc("/card/{card_number}/info", getCardFundsAPIHandler).Methods("GET")
	r.HandleFunc("/card/{card_number}/load_funds", loadFundsAPIHandler).Methods("POST")
	r.HandleFunc("/card/{card_number}/authorization_request", authRequestAPIHandler).Methods("POST")
	r.HandleFunc("/card/{card_number}/capture/{auth_request_uuid}", captureFundsAPIHandler).Methods("POST")
	r.HandleFunc("/card/{card_number}/authorization_request/{auth_request_uuid}/reverse", reverseAuthRqAPIHandler).Methods("POST")
	r.HandleFunc("/card/{card_number}/capture/{auth_request_uuid}/refund", refundFundsAPIHandler).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", r))
}

func createCardAPIHandler(w http.ResponseWriter, r *http.Request) {
	d := json.NewDecoder(r.Body)

	var data createCardRequest
	err := d.Decode(&data)
	if err != nil {
		log.Printf("Error decoding JSON. Error: %v", err)
		http.Error(w, errIncorrectOrMalformedJSON, http.StatusBadRequest)
		return
	}

	if !createCardRqIsValid(data) {
		http.Error(w, errRqInvalidOrMissingData, http.StatusBadRequest)
		return
	}

	prepaidCard, err := prepaid.NewCard(data.CardholderName, data.CardCurrency)
	if err != nil {
		log.Printf("Error creating card: Error: %v", err)
		http.Error(w, errCardCreation, http.StatusInternalServerError)
		return
	}

	resp := fmt.Sprintf(cardCreationSuccess, prepaidCard.CardNumber)
	w.Write([]byte(resp))
}

func getCardFundsAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	possibleCardNo, err := validateAndReturnCardNumber(vars)
	if err != nil {
		log.Printf("Error validating card number in request. Error: %v", err)
		http.Error(w, errInvalidCardNumProvided, http.StatusBadRequest)
		return
	}

	card, err := prepaid.FindCardByNumber(*possibleCardNo)
	if err != nil {
		http.Error(w, errCardNotFound, http.StatusNotFound)
		return
	}

	jsonResponse, err := json.Marshal(card)
	if err != nil {
		// This should never happen... But I put it here because it will, ofc
		log.Printf("Error marshalling JSON response. Error: %v", err)
		http.Error(w, errBackendGeneric, http.StatusInternalServerError)
		return
	}

	w.Write(jsonResponse)
}

func loadFundsAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	possibleCardNo, err := validateAndReturnCardNumber(vars)
	if err != nil {
		log.Printf("Error validating card number in request. Error: %v", err)
		http.Error(w, errInvalidCardNumProvided, http.StatusBadRequest)
		return
	}

	d := json.NewDecoder(r.Body)

	var data loadFundsRequest
	err = d.Decode(&data)
	if err != nil {
		log.Printf("Error decoding JSON. Error: %v", err)
		http.Error(w, errIncorrectOrMalformedJSON, http.StatusBadRequest)
		return
	}

	if !loadFundsRqIsValid(data) {
		http.Error(w, errRqInvalidOrMissingData, http.StatusBadRequest)
		return
	}

	card, err := prepaid.FindCardByNumber(*possibleCardNo)
	if err != nil {
		http.Error(w, errCardNotFound, http.StatusNotFound)
		return
	}

	card.LoadFunds(data.AmountCents)
}

func authRequestAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	cardNumber, err := validateAndReturnCardNumber(vars)
	if err != nil {
		log.Printf("Error validating card number in request. Error: %v", err)
		http.Error(w, errInvalidCardNumProvided, http.StatusBadRequest)
		return
	}

	d := json.NewDecoder(r.Body)

	var data authRequest
	err = d.Decode(&data)
	if err != nil {
		log.Printf("Error decoding JSON. Error: %v", err)
		http.Error(w, errIncorrectOrMalformedJSON, http.StatusBadRequest)
		return
	}

	if !authRqIsValid(data) {
		http.Error(w, errRqInvalidOrMissingData, http.StatusBadRequest)
		return
	}

	card, err := prepaid.FindCardByNumber(*cardNumber)
	if err != nil {
		http.Error(w, errCardNotFound, http.StatusNotFound)
		return
	}

	authRqUUID, err := card.NewAuthRequest(vars["merchant_name"], data.AuthAmount)
	if err != nil {
		log.Printf("Error creating auth request. Error: %v", err)
		if err == prepaid.ErrCardIsNotUsable {
			http.Error(w, errCardIsNotUsable, http.StatusForbidden)
			return
		}

		http.Error(w, errNotEnoughFunds, http.StatusForbidden)
		return
	}

	resp := fmt.Sprintf(authRequestSuccess, *authRqUUID)
	w.Write([]byte(resp))
}

func reverseAuthRqAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	authRqUUID := vars["auth_request_uuid"]
	if authRqUUID == "" {
		http.Error(w, errRqInvalidOrMissingData, http.StatusBadRequest)
		return
	}

	cardNumber, err := validateAndReturnCardNumber(vars)
	if err != nil {
		log.Printf("Error validating card number in request. Error: %v", err)
		http.Error(w, errInvalidCardNumProvided, http.StatusBadRequest)
		return
	}

	d := json.NewDecoder(r.Body)

	var data reverseAuthRequest
	err = d.Decode(&data)
	if err != nil {
		log.Printf("Error decoding JSON. Error: %v", err)
		http.Error(w, errIncorrectOrMalformedJSON, http.StatusBadRequest)
		return
	}

	if !reverseAuthRqIsValid(data) {
		http.Error(w, errRqInvalidOrMissingData, http.StatusBadRequest)
		return
	}

	card, err := prepaid.FindCardByNumber(*cardNumber)
	if err != nil {
		http.Error(w, errCardNotFound, http.StatusNotFound)
		return
	}

	authRq, err := prepaid.FindAuthRequestByUUID(card.CardNumber, authRqUUID)
	if err != nil {
		log.Printf("Error obtaining Auth Request. Error: %v", err)
		http.Error(w, errAuthRqNotFound, http.StatusNotFound)
		return
	}

	if data.ReverseAmount > authRq.AuthAmount {
		// You can't reverse more than you were authorized for!
		http.Error(w, errInvalidReverseAmount, http.StatusForbidden)
		return
	}

	fundCapture, err := prepaid.FindFundCaptureByAuthUUID(authRq.UUID)
	if err == nil {
		if fundCapture.CapturedAmount > data.ReverseAmount {
			// You can't reverse an amount (or greater) which you've captured already!
			http.Error(w, errInvalidReverseAmount, http.StatusForbidden)
			return
		}
	}

	// From here on out, there's no fund capture means we don't have to worry with
	// the edge case above
	err = authRq.Reverse(data.ReverseAmount, card)
	if err != nil {
		http.Error(w, errCardIsNotUsable, http.StatusForbidden)
		return
	}
}

func captureFundsAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	authRqUUID := vars["auth_request_uuid"]
	if authRqUUID == "" {
		http.Error(w, errRqInvalidOrMissingData, http.StatusBadRequest)
		return
	}

	cardNumber, err := validateAndReturnCardNumber(vars)
	if err != nil {
		log.Printf("Error validating card number in request. Error: %v", err)
		http.Error(w, errInvalidCardNumProvided, http.StatusBadRequest)
		return
	}

	d := json.NewDecoder(r.Body)

	var data captureFundsRequest
	err = d.Decode(&data)
	if err != nil {
		log.Printf("Error decoding JSON. Error: %v", err)
		http.Error(w, errIncorrectOrMalformedJSON, http.StatusBadRequest)
		return
	}

	if !captureFundsRqIsValid(data) {
		http.Error(w, errRqInvalidOrMissingData, http.StatusBadRequest)
		return
	}

	card, err := prepaid.FindCardByNumber(*cardNumber)
	if err != nil {
		http.Error(w, errCardNotFound, http.StatusNotFound)
		return
	}

	authRq, err := prepaid.FindAuthRequestByUUID(card.CardNumber, authRqUUID)
	if err != nil {
		log.Printf("Error obtaining Auth Request. Error: %v", err)
		http.Error(w, errAuthRqNotFound, http.StatusNotFound)
		return
	}

	fundCapture, err := prepaid.FindFundCaptureByAuthUUID(authRq.UUID)
	if err != nil {
		// Fund capture does not exist yet
		fundCapture = prepaid.NewFundCapture(authRq)
	}

	err = fundCapture.Capture(data.CaptureAmount, card)
	if err != nil {
		log.Printf("Error capturing. Error: %v", err)
		if err == prepaid.ErrAuthIsReversed {
			http.Error(w, errAuthReversed, http.StatusForbidden)
			return
		}
		http.Error(w, errInvalidCaptureAmount, http.StatusBadRequest)
	}
}

func refundFundsAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	authRqUUID := vars["auth_request_uuid"]
	if authRqUUID == "" {
		http.Error(w, errRqInvalidOrMissingData, http.StatusBadRequest)
		return
	}

	cardNumber, err := validateAndReturnCardNumber(vars)
	if err != nil {
		log.Printf("Error validating card number in request. Error: %v", err)
		http.Error(w, errInvalidCardNumProvided, http.StatusBadRequest)
		return
	}

	d := json.NewDecoder(r.Body)

	var data refundFundsRequest
	err = d.Decode(&data)
	if err != nil {
		log.Printf("Error decoding JSON. Error: %v", err)
		http.Error(w, errIncorrectOrMalformedJSON, http.StatusBadRequest)
		return
	}

	if !refundFundsRqIsValid(data) {
		http.Error(w, errRqInvalidOrMissingData, http.StatusBadRequest)
		return
	}

	card, err := prepaid.FindCardByNumber(*cardNumber)
	if err != nil {
		http.Error(w, errCardNotFound, http.StatusNotFound)
		return
	}

	authRq, err := prepaid.FindAuthRequestByUUID(card.CardNumber, authRqUUID)
	if err != nil {
		log.Printf("Error obtaining Auth Request. Error: %v", err)
		http.Error(w, errAuthRqNotFound, http.StatusNotFound)
		return
	}

	fundCapture, err := prepaid.FindFundCaptureByAuthUUID(authRq.UUID)
	if err != nil {
		http.Error(w, errFundCaptureNotFound, http.StatusNotFound)
		return
	}

	if data.RefundAmount > fundCapture.CapturedAmount {
		// You can't refund more than you captured!
		http.Error(w, errInvalidRefundAmount, http.StatusForbidden)
		return
	}

	card.Refund(data.RefundAmount)
	fundCapture.CapturedAmount -= data.RefundAmount
}
