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

type AppScriptProvider struct {
	webAppURL  string
	secret     string
	httpClient *http.Client
}

func NewAppScriptProvider(webAppURL, secret string) *AppScriptProvider {
	return &AppScriptProvider{
		webAppURL: webAppURL,
		secret:    secret,
		httpClient: &http.Client{
			Timeout: 20 * time.Second, // Apps Script kadang agak lambat cold-start
		},
	}
}

type appScriptRequestBody struct {
	Secret      string `json:"secret"`
	ToEmail     string `json:"to_email"`
	ToName      string `json:"to_name,omitempty"`
	Subject     string `json:"subject"`
	HTMLContent string `json:"html_content,omitempty"`
	TextContent string `json:"text_content,omitempty"`
}

type appScriptResponseBody struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	MessageID string `json:"message_id"`
}

func (p *AppScriptProvider) SendEmail(ctx context.Context, in SendEmailInput) (SendEmailOutput, error) {
	if in.ToEmail == "" {
		return SendEmailOutput{}, fmt.Errorf("to_email wajib diisi")
	}
	if in.Subject == "" {
		return SendEmailOutput{}, fmt.Errorf("subject wajib diisi")
	}

	body := appScriptRequestBody{
		Secret:      p.secret,
		ToEmail:     in.ToEmail,
		ToName:      in.ToName,
		Subject:     in.Subject,
		HTMLContent: in.HTMLContent,
		TextContent: in.TextContent,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return SendEmailOutput{}, fmt.Errorf("marshal payload apps script: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.webAppURL, bytes.NewReader(payload))
	if err != nil {
		return SendEmailOutput{}, fmt.Errorf("build request apps script: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return SendEmailOutput{}, fmt.Errorf("hubungi apps script: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return SendEmailOutput{}, fmt.Errorf("baca response apps script: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return SendEmailOutput{}, fmt.Errorf("apps script status %d: %s", resp.StatusCode, string(rawBody))
	}

	var parsed appScriptResponseBody
	if err := json.Unmarshal(rawBody, &parsed); err != nil {
		return SendEmailOutput{}, fmt.Errorf("parse response apps script: %w (raw: %s)", err, string(rawBody))
	}

	if !parsed.Success {
		return SendEmailOutput{}, fmt.Errorf("apps script gagal kirim: %s", parsed.Message)
	}

	return SendEmailOutput{MessageID: parsed.MessageID}, nil
}
