package api

import (
	"strconv"

	"github.com/hugoamvieira/prepaid-card-project/backend/src/prepaid"
	"github.com/pkg/errors"
)

func createCardRqIsValid(data createCardRequest) bool {
	// A `/card/new` rq is valid if the currency is in the map and the cardholder's name is not empty.
	_, isValidCurrency := prepaid.SupportedCurrencies[data.CardCurrency]
	return isValidCurrency && data.CardholderName != ""
}

func loadFundsRqIsValid(data loadFundsRequest) bool {
	// A `load_funds` rq is valid if the amount in cents is greater than 0.
	return data.AmountCents > 0
}

func authRqIsValid(data authRequest) bool {
	// An auth request is valid if the amount in cents is greater than 0 and merchant name is not empty.
	return data.AuthAmount > 0 && data.MerchantName != ""
}

func captureFundsRqIsValid(data captureFundsRequest) bool {
	// A capture funds  is valid if the amount to capture is greater than 0.
	return data.CaptureAmount > 0
}

func reverseAuthRqIsValid(data reverseAuthRequest) bool {
	// A reverse auth rq is valid if the amount to reverse is greater than 0.
	return data.ReverseAmount > 0
}

func refundFundsRqIsValid(data refundFundsRequest) bool {
	// A refund rq is valid if the amount to refund is greater than 0.
	return data.RefundAmount > 0
}

func validateAndReturnCardNumber(vars map[string]string) (*int64, error) {
	cardNumberString, exists := vars["card_number"]
	if !exists {
		return nil, errors.New("card number was not specified")
	}

	cardNumber, err := strconv.ParseInt(cardNumberString, 10, 64)
	if err != nil {
		errors.Wrap(err, "error parsing card number to int64")
		return nil, err
	}

	return &cardNumber, nil
}
