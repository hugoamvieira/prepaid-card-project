package prepaid

import (
	"errors"
)

var (
	fundCaptures = map[string]*FundCapture{} // AuthRq UUID -> *FundCapture

	// Errors
	ErrAuthIsReversed = errors.New("Authorization Request has been reversed")
	ErrInvalidCapture = errors.New("Invalid capture amount")

	errNoFundCapture = errors.New("No fund capture for that auth request")
)

type FundCapture struct {
	AuthRequest    *AuthRequest
	CapturedAmount int64 // In cents
	Completed      bool
}

func NewFundCapture(authRq *AuthRequest) *FundCapture {
	fc := &FundCapture{
		AuthRequest:    authRq,
		CapturedAmount: 0,
		Completed:      false,
	}

	fundCaptures[authRq.UUID] = fc
	return fc
}

func (fc *FundCapture) Capture(amount int64, card *Card) error {
	if fc.Completed {
		return ErrInvalidCapture
	}
	if fc.CapturedAmount+amount > fc.AuthRequest.AuthAmount {
		return ErrInvalidCapture
	}
	if fc.CapturedAmount+amount == fc.AuthRequest.AuthAmount {
		if !fc.Completed {
			fc.Completed = true
		}
	}

	fc.CapturedAmount += amount
	err := card.CaptureFunds(amount)
	if err != nil {
		return err
	}

	return nil
}

func FindFundCaptureByAuthUUID(authUUID string) (*FundCapture, error) {
	fc, exists := fundCaptures[authUUID]
	if !exists {
		return nil, errNoFundCapture
	}
	return fc, nil
}
