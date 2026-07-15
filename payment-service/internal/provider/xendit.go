package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ____________________________________________________________________________
// Konstanta
// ____________________________________________________________________________

// Nilai-nilai berikut DITUNTUT oleh protokol Xendit, bukan pilihan kita.
// Sengaja TIDAK dijadikan config: kalau diubah, request-nya jadi invalid dan
// Xendit menolak. Config hanya untuk nilai yang aman diubah antar-lingkungan.
const (
	sessionTypePay     = "PAY"          // sesi untuk menerima pembayaran
	modePaymentLink    = "PAYMENT_LINK" // Xendit yang meng-host halaman checkout
	customerIndividual = "INDIVIDUAL"   // tipe customer perorangan (bukan bisnis)

	// Status sesi dari Xendit.
	xenditStatusActive    = "ACTIVE"    // sesi hidup, belum dibayar
	xenditStatusCompleted = "COMPLETED" // sudah dibayar
	xenditStatusExpired   = "EXPIRED"   // lewat batas waktu
	xenditStatusCanceled  = "CANCELED"  // dibatalkan

	pathSessions = "/sessions"
)

// Default yang dipakai kalau config tidak mengisinya. Ini boleh diubah
// lewat config — beda dengan konstanta protokol di atas.
const (
	defaultXenditBaseURL     = "https://api.xendit.co"
	defaultXenditHTTPTimeout = 15 * time.Second
	defaultCurrency          = "IDR"
	defaultCountry           = "ID"
)

// ____________________________________________________________________________
// Konstruksi
// ____________________________________________________________________________

// XenditOptions dioper dari factory (yang mengambilnya dari config).
// Pakai struct, bukan parameter berderet, supaya penambahan opsi baru
// tidak memaksa semua pemanggil berubah.
type XenditOptions struct {
	APIKey      string        // secret key (xnd_development_... / xnd_production_...)
	BaseURL     string        // kosong -> defaultXenditBaseURL
	ReturnURL   string        // tujuan redirect user setelah bayar. WAJIB HTTPS.
	HTTPTimeout time.Duration // kosong -> defaultXenditHTTPTimeout
	Currency    string        // kosong -> defaultCurrency
	Country     string        // kosong -> defaultCountry
}

// Xendit adalah PaymentProvider yang berbicara ke Xendit Payment Session API.
//
// Stateless: tidak menyimpan apa pun. Sumber kebenaran ada di Xendit (untuk
// status sesi) dan di core-service (untuk order). Karena itu korelasi balik ke
// order dilakukan lewat reference_id, bukan lewat penyimpanan lokal.
type Xendit struct {
	apiKey     string
	baseURL    string
	returnURL  string
	currency   string
	country    string
	httpClient *http.Client
}

// NewXendit mengisi default untuk opsi yang kosong, supaya env yang belum
// diisi tidak diam-diam mengirim field kosong dan ditolak Xendit.
func NewXendit(o XenditOptions) *Xendit {
	if o.BaseURL == "" {
		o.BaseURL = defaultXenditBaseURL
	}
	if o.HTTPTimeout == 0 {
		o.HTTPTimeout = defaultXenditHTTPTimeout
	}
	if o.Currency == "" {
		o.Currency = defaultCurrency
	}
	if o.Country == "" {
		o.Country = defaultCountry
	}

	return &Xendit{
		apiKey:     o.APIKey,
		baseURL:    o.BaseURL,
		returnURL:  o.ReturnURL,
		currency:   o.Currency,
		country:    o.Country,
		httpClient: &http.Client{Timeout: o.HTTPTimeout},
	}
}

// Pastikan Xendit memenuhi kontrak PaymentProvider saat kompilasi.
var _ PaymentProvider = (*Xendit)(nil)

// ____________________________________________________________________________
// Implementasi PaymentProvider
// ____________________________________________________________________________

// CreateInvoice membuat payment session dan mengembalikan link checkout.
//
// Catatan penting: p.Amount adalah OTORITATIF. Total tidak pernah dihitung
// ulang dari p.Items — perhitungan harga adalah tanggung jawab core-service.
// Items hanya diteruskan apa adanya untuk tampilan di halaman checkout.
func (x *Xendit) CreateInvoice(ctx context.Context, p CreateInvoiceParams) (Invoice, error) {
	body := sessionRequest{
		ReferenceID: referenceFromOrderID(p.OrderID),
		SessionType: sessionTypePay,
		Mode:        modePaymentLink,
		Amount:      p.Amount,
		Currency:    x.currency,
		Country:     x.country,
		Customer: customer{
			ReferenceID:      fmt.Sprintf("user-%d", p.UserID),
			Type:             customerIndividual,
			Email:            p.CustomerEmail,
			IndividualDetail: individualDetail{GivenNames: p.CustomerName},
		},
		Items:            toWireItems(p.Items),
		SuccessReturnURL: x.returnURL,
		CancelReturnURL:  x.returnURL,
	}

	var resp sessionResponse
	if err := x.do(ctx, http.MethodPost, pathSessions, body, &resp); err != nil {
		return Invoice{}, err
	}

	return Invoice{
		Reference:   resp.PaymentSessionID,
		PaymentLink: resp.PaymentLinkURL,
		Status:      toProviderStatus(resp.Status),
		ExpiresAt:   resp.ExpiresAt,
		OrderID:     p.OrderID, // sudah diketahui pemanggil, tak perlu parsing
	}, nil
}

// GetInvoice mengambil status sesi terkini dari Xendit.
//
// Inilah cara core mengetahui pembayaran sudah masuk: tidak ada webhook,
// jadi core mem-polling method ini sampai status berubah dari PENDING.
func (x *Xendit) GetInvoice(ctx context.Context, reference string) (Invoice, error) {
	var resp sessionResponse
	if err := x.do(ctx, http.MethodGet, pathSessions+"/"+reference, nil, &resp); err != nil {
		return Invoice{}, err
	}

	// Sesi yang tidak dibuat oleh sistem kita bisa punya reference_id dengan
	// format lain. Dalam kasus itu OrderID jadi 0 — bukan error, karena core
	// selalu memanggil dengan reference yang sudah dia simpan sendiri.
	orderID := orderIDFromReference(resp.ReferenceID)

	return Invoice{
		Reference:   resp.PaymentSessionID,
		PaymentLink: resp.PaymentLinkURL,
		Status:      toProviderStatus(resp.Status),
		ExpiresAt:   resp.ExpiresAt,
		OrderID:     orderID,
	}, nil
}

// ____________________________________________________________________________
// Kontrak wire Xendit
//
// Struct di bawah ini adalah bentuk JSON milik Xendit, bukan tipe domain kita.
// Sengaja tidak diekspor: dunia luar hanya boleh melihat tipe netral
// (Invoice, CreateInvoiceParams) — itulah gunanya provider sebagai lapisan
// penerjemah. Hanya field yang benar-benar dipakai yang didefinisikan.
// ____________________________________________________________________________

type sessionRequest struct {
	ReferenceID      string     `json:"reference_id"`
	SessionType      string     `json:"session_type"`
	Mode             string     `json:"mode"`
	Amount           int64      `json:"amount"`
	Currency         string     `json:"currency"`
	Country          string     `json:"country"`
	Customer         customer   `json:"customer"`
	Items            []wireItem `json:"items,omitempty"`
	SuccessReturnURL string     `json:"success_return_url,omitempty"`
	CancelReturnURL  string     `json:"cancel_return_url,omitempty"`
}

type customer struct {
	ReferenceID      string           `json:"reference_id"`
	Type             string           `json:"type"`
	Email            string           `json:"email"`
	IndividualDetail individualDetail `json:"individual_detail"`
}

type individualDetail struct {
	GivenNames string `json:"given_names"`
}

type wireItem struct {
	Name     string `json:"name"`
	Quantity int32  `json:"quantity"`
	Price    int64  `json:"price"`
}

type sessionResponse struct {
	PaymentSessionID string    `json:"payment_session_id"`
	ReferenceID      string    `json:"reference_id"`
	Status           string    `json:"status"`
	Amount           int64     `json:"amount"`
	PaymentLinkURL   string    `json:"payment_link_url"`
	ExpiresAt        time.Time `json:"expires_at"`
}

// HTTP plumbing

// do menangani satu request ke Xendit: encode body, pasang auth, cek status,
// decode respons. Semua method di atas lewat sini supaya penanganan error dan
// autentikasi cuma ditulis di satu tempat.
//
// in boleh nil (untuk GET). out boleh nil kalau respons tidak dipakai.
func (x *Xendit) do(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		encoded, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("xendit: failed to encode request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, x.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("xendit: failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Xendit memakai Basic auth: secret key sebagai username, password kosong.
	req.SetBasicAuth(x.apiKey, "")

	res, err := x.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("xendit: request failed: %w", err)
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("xendit: failed to read response: %w", err)
	}

	// 404 dipetakan ke sentinel error supaya grpcserver bisa menerjemahkannya
	// jadi codes.NotFound tanpa perlu tahu ini provider yang mana.
	if res.StatusCode == http.StatusNotFound {
		return ErrInvoiceNotFound
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		// Body Xendit disertakan karena di situlah nama field yang bermasalah
		// disebut. Error ini hanya masuk log server, tidak dibocorkan ke core.
		return fmt.Errorf("xendit: status %d: %s", res.StatusCode, raw)
	}

	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("xendit: failed to decode response: %w", err)
		}
	}
	return nil
}

// Konversi antara bahasa Xendit dan bahasa provider

// referenceFromOrderID dan orderIDFromReference adalah sepasang.
// reference_id inilah yang membuat provider ini bisa stateless: order_id kita
// dititipkan ke Xendit, lalu dibaca lagi saat status ditanyakan.
const orderRefPrefix = "order-"

func referenceFromOrderID(orderID int64) string {
	return fmt.Sprintf("%s%d", orderRefPrefix, orderID)
}

// Mengembalikan 0 kalau format tidak dikenali. Lihat catatan di GetInvoice.
func orderIDFromReference(ref string) int64 {
	var id int64
	if _, err := fmt.Sscanf(ref, orderRefPrefix+"%d", &id); err != nil {
		return 0
	}
	return id
}

func toProviderStatus(s string) Status {
	switch s {
	case xenditStatusActive:
		return StatusPending
	case xenditStatusCompleted:
		return StatusPaid
	case xenditStatusExpired:
		return StatusExpired
	case xenditStatusCanceled:
		return StatusFailed
	default:
		// Status yang tidak dikenal diperlakukan sebagai belum lunas —
		// lebih aman menganggap belum bayar daripada salah menganggap lunas.
		return StatusPending
	}
}

func toWireItems(items []InvoiceItem) []wireItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]wireItem, 0, len(items))
	for _, it := range items {
		out = append(out, wireItem{Name: it.Name, Quantity: it.Quantity, Price: it.Price})
	}
	return out
}
