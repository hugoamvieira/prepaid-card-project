package prepaid

import (
	"errors"
	"math/rand"
)

const (
	createCardMaxRetries = 5
)

var (
	// Map of Card Number to `Card`. This is the source of truth for cards
	cards = map[int64]*Card{}

	// SupportedCurrencies is a map that has all the supported currencies for this system
	SupportedCurrencies = map[string]string{
		"GBP": "Pound Sterling",
		"USD": "United States Dollar",
	}

	// ErrCardIsNotUsable is returned when the card has not been loaded with money
	ErrCardIsNotUsable = errors.New("the card is not usable")

	// ErrNotEnoughFunds is returned when there's not enough funds in the card for an auth request
	ErrNotEnoughFunds = errors.New("the card doesn't have enough available funds")
)

// Card (in the Prepaid namespace) represents what would be in the database regarding card information.
type Card struct {
	CardNumber   int64
	Cardholder   Cardholder
	CardCurrency string // Must be one of `supportedCurrencies`
	LoadedFunds  int64  `json:"loaded_funds"`  // In Cents
	BlockedFunds int64  `json:"blocked_funds"` // In Cents
	TotalFunds   int64  `json:"total_funds"`   // In Cents
	Usable       bool
}

// Cardholder represents the cardholder's info.
type Cardholder struct {
	Name string
}

// NewCard tries to create a new prepaid card and returns the card object if successful.
func NewCard(cardholderName string, chosenCurrency string) (*Card, error) {
	var cNum int64
	var cNumUnique bool
	for i := 0; i < createCardMaxRetries; i++ {
		cNum = rand.Int63()
		_, exists := cards[cNum]
		if !exists {
			cNumUnique = true
			break
		}
	}

	if !cNumUnique {
		return nil, errors.New("Couldn't find a unique card number")
	}

	cHolder := Cardholder{
		Name: cardholderName,
	}

	pc := &Card{
		CardNumber:   cNum,
		Cardholder:   cHolder,
		CardCurrency: chosenCurrency,
		LoadedFunds:  0,
		BlockedFunds: 0,
		TotalFunds:   0,
		Usable:       false,
	}

	cards[pc.CardNumber] = pc

	return pc, nil
}

// FindCardByNumber finds the card object associated with a certain card number.
func FindCardByNumber(cardNumber int64) (*Card, error) {
	c, exists := cards[cardNumber]
	if !exists {
		return nil, errors.New("Card with that number does not exist")
	}
	return c, nil
}

// GetCardsList returns a list of all the registered prepaid cards
func GetCardsList() map[int64]*Card {
	return cards
}

// LoadFunds loads money into a certain card. It also makes it usable
func (c *Card) LoadFunds(amountCents int64) {
	c.LoadedFunds += amountCents
	c.TotalFunds += amountCents
	c.Usable = true
}

func (c *Card) Refund(amountCents int64) {
	c.TotalFunds += amountCents
}

// BlockFunds will block funds on a card if there's enough money available and if it's usable.
func (c *Card) BlockFunds(amountCents int64) error {
	if !c.Usable {
		return ErrCardIsNotUsable
	}
	if !c.hasEnoughAvailableFunds(amountCents) {
		return ErrNotEnoughFunds
	}

	c.BlockedFunds += amountCents
	return nil
}

func (c *Card) UnblockFunds(amountCents int64) error {
	if !c.Usable {
		return ErrCardIsNotUsable
	}
	c.BlockedFunds -= amountCents
	return nil
}

func (c *Card) CaptureFunds(amountCents int64) error {
	if amountCents > c.BlockedFunds {
		return ErrNotEnoughFunds
	}

	c.BlockedFunds -= amountCents
	c.TotalFunds -= amountCents
	return nil
}

func (c *Card) hasEnoughAvailableFunds(amountCents int64) bool {
	return c.TotalFunds-c.BlockedFunds >= amountCents
}
