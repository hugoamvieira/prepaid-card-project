package prepaid

import (
	"errors"

	"github.com/satori/go.uuid"
)

var (
	// Source of truth for auth requests
	authRequests = map[int64][]*AuthRequest{}

	// Errors
	errAuthRqDoesNotExist = errors.New("Authorization Request does not exist.")
)

// AuthRequest represents an authorization request by a merchant
type AuthRequest struct {
	UUID         string
	MerchantName string
	AuthAmount   int64 // In Cents
}

// NewAuthRequest creates a new autorization request for funds. Returns the UUID for the Auth Request or error if not enough funds.
func (c *Card) NewAuthRequest(merchantName string, authAmount int64) (*string, error) {
	ar := &AuthRequest{
		UUID:         uuid.NewV4().String(),
		MerchantName: merchantName,
		AuthAmount:   authAmount,
	}

	err := c.BlockFunds(authAmount)
	if err != nil {
		// Error can either be "card not usable" or "not enough funds"
		return nil, err
	}

	authRequests[c.CardNumber] = append(authRequests[c.CardNumber], ar)
	return &ar.UUID, nil
}

func (ar *AuthRequest) Reverse(amount int64, card *Card) error {
	ar.AuthAmount -= amount
	err := card.UnblockFunds(amount)
	if err != nil {
		return err
	}
	return nil
}

// FindAuthRequestByUUID will try to find the requested auth rq based on cardNumber and UUID
func FindAuthRequestByUUID(cardNumber int64, uuid string) (*AuthRequest, error) {
	authRqs, exists := authRequests[cardNumber]
	if !exists {
		return nil, errAuthRqDoesNotExist
	}

	for _, arq := range authRqs {
		if uuid == arq.UUID {
			return arq, nil
		}
	}

	return nil, errAuthRqDoesNotExist
}
