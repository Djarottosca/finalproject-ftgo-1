package provider

import (
	"fmt"
	"time"
)

type Config struct {
	Name       string
	BaseURL    string
	InvoiceTTL time.Duration

	XenditAPIKey      string
	XenditBaseURL     string
	XenditReturnURL   string
	XenditHTTPTimeout time.Duration
	XenditCurrency    string
	XenditCountry     string
}

func New(cfg Config) (PaymentProvider, error) {
	switch cfg.Name {
	case "simulation":
		return NewSimulation(cfg.BaseURL, cfg.InvoiceTTL), nil
	case "xendit":
		return NewXendit(XenditOptions{
			APIKey:      cfg.XenditAPIKey,
			BaseURL:     cfg.XenditBaseURL,
			ReturnURL:   cfg.XenditReturnURL,
			HTTPTimeout: cfg.XenditHTTPTimeout,
			Currency:    cfg.XenditCurrency,
			Country:     cfg.XenditCountry,
		}), nil
	default:
		return nil, fmt.Errorf("provider: nama provider tidak dikenal: %q", cfg.Name)
	}
}
