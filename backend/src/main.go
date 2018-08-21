package main

import (
	"fmt"
	"time"

	"github.com/hugoamvieira/prepaid-card-project/backend/src/api"
	"github.com/hugoamvieira/prepaid-card-project/backend/src/prepaid"
)

func main() {
	finished := make(chan bool, 1)

	go api.NewAPI()

	go func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				a := prepaid.GetCardsList()
				for _, v := range a {
					fmt.Printf("Card Number %v (%v):\nLoaded Funds: %v\nBlocked Funds: %v\nTotal Funds: %v\nCardholder: %v\nIs Usable?: %v\n\n",
						v.CardNumber, v.CardCurrency, v.LoadedFunds, v.BlockedFunds, v.TotalFunds, v.Cardholder.Name, v.Usable)
				}
			}
		}
	}()

	<-finished
}
