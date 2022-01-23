package main

import (
	"log"
	"net/http"
	"server/handler"
	"server/utils"

	"github.com/go-chi/chi"
	"github.com/stripe/stripe-go/v72"
)

type Application struct {
	Environment string
	StripeKey   string
}

func main() {
	app := Application{}
	app.Environment = utils.GetEnvOrDefault("ENV", "local")
	app.StripeKey = utils.GetEnvOrDefault("STRIPE_KEY", "sk_test_4eC39HqLyjWDarjtT1zdp7dc")

	stripe.Key = app.StripeKey

	r := chi.NewRouter()

	// Middlewares
	// r.Use(middleware.Logger)

	// Handlers
	r.Get("/health", handler.Health)
	r.Post("/create-checkout-session", handler.CreateCheckoutSession)
	r.Post("/create-portal-session", handler.CreatePortalSession)
	r.Post("/webhook", handler.Webhook)

	addr := "localhost:3000"
	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
