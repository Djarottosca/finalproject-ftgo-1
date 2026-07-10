package provider

import "fmt"

// Config: cuma yang provider butuh. Sengaja kecil biar package provider
// gak kenal Viper / config aplikasi. main.go yang mapping dari app config ke sini.
type Config struct {
	Name                string // "simulation" | "xendit"
	BaseURL             string // buat simulasi nyusun payment_link
	XenditAPIKey        string // kepakai nanti pas xendit.go jadi
	XenditCallbackToken string // kepakai nanti pas xendit.go jadi
}

// New milih implementasi berdasar cfg.Name. Balikannya interface, jadi
// pemanggil (main) gak kenal tipe konkret.
func New(cfg Config) (PaymentProvider, error) {
	switch cfg.Name {
	case "simulation":
		return NewSimulation(cfg.BaseURL), nil
	case "xendit":
		// return NewXendit(cfg.XenditAPIKey, cfg.XenditCallbackToken), nil
		return nil, fmt.Errorf("provider: xendit belum diimplementasi")
	default:
		return nil, fmt.Errorf("provider: nama provider tidak dikenal: %q", cfg.Name)
	}
}
