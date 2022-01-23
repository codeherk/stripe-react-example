package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v72"
	portalsession "github.com/stripe/stripe-go/v72/billingportal/session"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/price"
	"github.com/stripe/stripe-go/webhook"
)

type ErrorResponse struct {
	message string
}

func Health(w http.ResponseWriter, r *http.Request) {
	log.Printf("here")
	w.WriteHeader(http.StatusOK)
	return
}

func CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	// if r.Method != "POST" {
	// 	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	// 	return
	// }
	log.Printf("Here")
	r.ParseForm()
	log.Printf("REQUESTFORM=%+v\n", r.Form)
	lookup_key := r.PostFormValue("lookup_key")
	log.Printf("lookup_key=%+v\n", lookup_key)

	domain := "http://localhost:3100"
	params := &stripe.PriceListParams{
		LookupKeys: stripe.StringSlice([]string{
			lookup_key,
		}),
	}

	log.Printf("Here 2")
	i := price.List(params)
	log.Printf("i : %+v", i)

	var price *stripe.Price

	for i.Next() {
		p := i.Price()
		log.Printf("p : %+v", p)
		price = p
	}
	if price == nil {
		// errRes := ErrorResponse{message: "No price found"}
		// w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		// json.NewEncoder(w).Encode(data)
		// http.(w, "No price found", http.StatusInternalServerError)
		return
	}
	log.Printf("Here 2.5")
	checkoutParams := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(price.ID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(domain + "?success=true&session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(domain + "?canceled=true"),
	}
	log.Printf("Here 3")
	s, err := session.New(checkoutParams)
	if err != nil {
		log.Printf("session.New: %v", err)
	}
	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func CreatePortalSession(w http.ResponseWriter, r *http.Request) {
	domain := "http://localhost:3100"
	// For demonstration purposes, we're using the Checkout session to retrieve the customer ID.
	// Typically this is stored alongside the authenticated user in your database.
	r.ParseForm()
	sessionId := r.PostFormValue("session_id")

	fmt.Print(sessionId)
	s, err := session.Get(sessionId, nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("session.Get: %v", err)
		return
	}

	// Authenticate your user.
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(s.Customer.ID),
		ReturnURL: stripe.String(domain),
	}
	ps, _ := portalsession.New(params)
	log.Printf("ps.New: %v", ps.URL)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ps.New: %v", err)
		return
	}

	http.Redirect(w, r, ps.URL, http.StatusSeeOther)
}

func Webhook(w http.ResponseWriter, req *http.Request) {
	const MaxBodyBytes = int64(65536)
	bodyReader := http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	// Replace this endpoint secret with your endpoint's unique secret
	// If you are testing with the CLI, find the secret by running 'stripe listen'
	// If you are using an endpoint defined with the API or dashboard, look in your webhook settings
	// at https://dashboard.stripe.com/webhooks
	endpointSecret := "whsec_123"
	signatureHeader := req.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signatureHeader, endpointSecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Webhook signature verification failed. %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}
	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "customer.subscription.deleted":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Subscription deleted for %s.", subscription.ID)
		// Then define and call a func to handle the deleted subscription.
		// handleSubscriptionCanceled(subscription)
	case "customer.subscription.updated":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Subscription updated for %s.", subscription.ID)
		// Then define and call a func to handle the successful attachment of a PaymentMethod.
		// handleSubscriptionUpdated(subscription)
	case "customer.subscription.created":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Subscription created for %s.", subscription.ID)
		// Then define and call a func to handle the successful attachment of a PaymentMethod.
		// handleSubscriptionCreated(subscription)
	case "customer.subscription.trial_will_end":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Subscription trial will end for %s.", subscription.ID)
		// Then define and call a func to handle the successful attachment of a PaymentMethod.
		// handleSubscriptionTrialWillEnd(subscription)
	default:
		fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
	}
	w.WriteHeader(http.StatusOK)
}
