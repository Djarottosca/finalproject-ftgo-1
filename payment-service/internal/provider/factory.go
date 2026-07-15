package provider

import "fmt"

// Config: cuma yang provider butuh. Sengaja kecil biar package provider
// gak kenal Viper / config aplikasi. main.go yang mapping dari app config ke sini.
type Config struct {
	Name            string
	BaseURL         string
	XenditAPIKey    string
	XenditBaseURL   string // default https://api.xendit.co
	XenditReturnURL string // halaman core tempat user dibalikin abis bayar
}

// New milih implementasi berdasar cfg.Name. Balikannya interface, jadi
// pemanggil (main) gak kenal tipe konkret.
func New(cfg Config) (PaymentProvider, error) {
	switch cfg.Name {
	case "simulation":
		return NewSimulation(cfg.BaseURL), nil
	case "xendit":
		return NewXendit(cfg.XenditAPIKey, cfg.XenditBaseURL, cfg.XenditReturnURL), nil
	default:
		return nil, fmt.Errorf("provider: nama provider tidak dikenal: %q", cfg.Name)
	}
}
