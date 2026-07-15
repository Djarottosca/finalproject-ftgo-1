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

const defaultXenditBaseURL = "https://api.xendit.co"

// Xendit adalah PaymentProvider yang nembak Xendit Payment Session API.
// Stateless: gak nyimpen apa-apa, semua state ada di Xendit.
type Xendit struct {
	apiKey     string
	baseURL    string // https://api.xendit.co
	returnURL  string // ke mana user dibalikin abis bayar (halaman core)
	httpClient *http.Client
}

func NewXendit(apiKey, baseURL, returnURL string) *Xendit {
	if baseURL == "" {
		baseURL = defaultXenditBaseURL
	}
	return &Xendit{
		apiKey:    apiKey,
		baseURL:   baseURL,
		returnURL: returnURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second, // jangan biarin request gantung selamanya
		},
	}
}

var _ PaymentProvider = (*Xendit)(nil)

// --- bentuk payload Xendit (cuma field yang kita butuh) ---

type xenditSessionRequest struct {
	ReferenceID     string              `json:"reference_id"`
	SessionType     string              `json:"session_type"` // "PAY"
	Mode            string              `json:"mode"`         // "PAYMENT_LINK"
	Amount          int64               `json:"amount"`
	Currency        string              `json:"currency"` // "IDR"
	Country         string              `json:"country"`  // "ID"
	Customer        xenditCustomer      `json:"customer"`
	Items           []xenditItem        `json:"items,omitempty"`
	SuccessReturnURL string             `json:"success_return_url,omitempty"`
	CancelReturnURL  string             `json:"cancel_return_url,omitempty"`
}

type xenditCustomer struct {
	ReferenceID      string                   `json:"reference_id"`
	Type             string                   `json:"type"` // "INDIVIDUAL"
	Email            string                   `json:"email"`
	IndividualDetail xenditIndividualDetail   `json:"individual_detail"`
}

type xenditIndividualDetail struct {
	GivenNames string `json:"given_names"`
}

type xenditItem struct {
	Name     string `json:"name"`
	Quantity int32  `json:"quantity"`
	Price    int64  `json:"price"`
}

type xenditSessionResponse struct {
	PaymentSessionID string    `json:"payment_session_id"`
	ReferenceID      string    `json:"reference_id"`
	Status           string    `json:"status"` // ACTIVE | COMPLETED | EXPIRED | CANCELED
	Amount           int64     `json:"amount"`
	PaymentLinkURL   string    `json:"payment_link_url"`
	ExpiresAt        time.Time `json:"expires_at"`
}

// --- implementasi interface ---

func (x *Xendit) CreateInvoice(ctx context.Context, p CreateInvoiceParams) (Invoice, error) {
	body := xenditSessionRequest{
		ReferenceID: fmt.Sprintf("order-%d", p.OrderID), // korelasi balik ke order kita
		SessionType: "PAY",
		Mode:        "PAYMENT_LINK",
		Amount:      p.Amount, // OTORITATIF, gak dihitung ulang dari items
		Currency:    "IDR",
		Country:     "ID",
		Customer: xenditCustomer{
			ReferenceID:      fmt.Sprintf("user-%d", p.UserID),
			Type:             "INDIVIDUAL",
			Email:            p.CustomerEmail,
			IndividualDetail: xenditIndividualDetail{GivenNames: p.CustomerName},
		},
		Items:            itemsToXendit(p.Items),
		SuccessReturnURL: x.returnURL,
		CancelReturnURL:  x.returnURL,
	}

	var resp xenditSessionResponse
	if err := x.do(ctx, http.MethodPost, "/sessions", body, &resp); err != nil {
		return Invoice{}, err
	}

	return Invoice{
		Reference:   resp.PaymentSessionID,
		PaymentLink: resp.PaymentLinkURL,
		Status:      xenditStatusToProvider(resp.Status),
		ExpiresAt:   resp.ExpiresAt,
		OrderID:     p.OrderID,
	}, nil
}

func (x *Xendit) GetInvoice(ctx context.Context, reference string) (Invoice, error) {
	var resp xenditSessionResponse
	if err := x.do(ctx, http.MethodGet, "/sessions/"+reference, nil, &resp); err != nil {
		return Invoice{}, err
	}

	orderID, _ := orderIDFromReference(resp.ReferenceID)

	return Invoice{
		Reference:   resp.PaymentSessionID,
		PaymentLink: resp.PaymentLinkURL,
		Status:      xenditStatusToProvider(resp.Status),
		ExpiresAt:   resp.ExpiresAt,
		OrderID:     orderID,
	}, nil
}

// --- plumbing HTTP ---

// do nembak Xendit: encode body, pasang Basic auth, decode respons.
func (x *Xendit) do(ctx context.Context, method, path string, in, out any) error {
	var reader io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("xendit: gagal encode request: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, x.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("xendit: gagal bikin request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Basic auth: secret key sebagai username, password kosong.
	req.SetBasicAuth(x.apiKey, "")

	res, err := x.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("xendit: request gagal: %w", err)
	}
	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("xendit: gagal baca respons: %w", err)
	}

	if res.StatusCode == http.StatusNotFound {
		return ErrInvoiceNotFound // dipetain ke codes.NotFound di grpcserver
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("xendit: status %d: %s", res.StatusCode, string(respBody))
	}

	if out != nil {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("xendit: gagal decode respons: %w", err)
		}
	}
	return nil
}

// --- mapping ---

func xenditStatusToProvider(s string) Status {
	switch s {
	case "ACTIVE":
		return StatusPending
	case "COMPLETED":
		return StatusPaid
	case "EXPIRED":
		return StatusExpired
	case "CANCELED":
		return StatusFailed
	default:
		return StatusPending
	}
}

func itemsToXendit(items []InvoiceItem) []xenditItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]xenditItem, 0, len(items))
	for _, it := range items {
		out = append(out, xenditItem{Name: it.Name, Quantity: it.Quantity, Price: it.Price})
	}
	return out
}

// orderIDFromReference ngebalikin "order-123" jadi 123.
func orderIDFromReference(ref string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(ref, "order-%d", &id)
	return id, err
}