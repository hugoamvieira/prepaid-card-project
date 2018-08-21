# Prepaid Card 

This simulates the workings of a very basic prepaid card.

## How to use it

- Run `dep ensure`
- Inside the `src` folder, run `go run *.go`
- You can cURL into the following endpoints:
  - 127.0.0.1:8000/card/new
    - Accepts POST
  - 127.0.0.1:8000/card/{card_number}/info
    - Accepts GET
  - 127.0.0.1:8000/card/{card_number}/load_funds
    - Accepts POST
  - 127.0.0.1:8000/card/{card_number}/authorization_request
      - Accepts POST
  - 127.0.0.1:8000/card/{card_number}/capture/{auth_request_uuid}
      - Accepts POST
  - 127.0.0.1:8000/card/{card_number}/authorization_request/{auth_request_uuid}/reverse
      - Accepts POST
  - 127.0.0.1:8000/card/{card_number}/capture/{auth_request_uuid}/refund
      - Accepts POST

Please see the examples below for JSON Body examples.

## Example cURL Requests

**Create New Card**

`curl 127.0.0.1:8000/card/new -d '{"cardholder_name": "Hugo Vieira", "card_currency": "GBP"}'`

**Get Card Info**

`curl 127.0.0.1:8000/card/<CARD_NUMBER>/info`

**Load Funds into Existing Card**

`curl 127.0.0.1:8000/card/<CARD_NUMBER>/load_funds -d '{"amount_cents": 50000}'`

**Merchant Auth Request**

`curl 127.0.0.1:8000/card/<CARD_NUMBER>/authorization_request -d '{"merchant_name": "Starbucks", "auth_amount": 520}'`

**Merchant Collect Request**

`curl 127.0.0.1:8000/card/<CARD_NUMBER>/capture/<AUTH_REQUEST_UUID> -d '{"capture_amount": 520}'`

**Merchant Reverse Auth Request**

`curl 127.0.0.1:8000/card/<CARD_NUMBER>/authorization_request/<AUTH_REQUEST_UUID>/reverse -d '{"reverse_amount": 100}'`

**Merchant Refund**

`curl 127.0.0.1:8000/card/<CARD_NUMBER>/capture/<AUTH_REQUEST_UUID>/refund -d '{"refund_amount": 100}'`

## Decision Log
- I have not written tests and the code is arguably a bit rushed - I really have no time and I'm trying to get this done in a reasonable amount of time;

- Card Verification (Card Checksumming, etc) is not considered - I feel like this was not relevant to the task;

- API to rule it all - This service has one API that deals with everything;

- The `frontend` component will periodically ping the `backend` component for info, making it not realtime... The shoemaker's son always goes barefoot;

- Looking up Prepaid Cards will a very frequent operation, so they will be stored in a map of card number to `prepaid.Card` struct - Makes it so that we have O(1) lookup time. Same for the `supportedCurrencies` map. Added bonus that we can have the complete currency name;

- API logic (In `api` package. Performs validation, etc) is conceptually and physically separated from the actual card lower level functions (In `prepaid` package) - Helps separate concerns and keep the code tidy and testable (even though it is not);

- The source of truth for Authorization Requests will be a map of card number to a slice of `prepaid.AuthRequest` - This ensures I can obtain all auth requests for one prepaid card.

- Each auth request is identified by a UUID, which is given to the merchant so they can come back to collect when they want;
