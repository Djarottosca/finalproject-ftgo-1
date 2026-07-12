package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Djarottosca/finalproject-ftgo-1/notification-service/internal/config"
)

const mailjetSendURL = "https://api.mailjet.com/v3.1/send"

type EmailSender interface {
	SendEmail(ctx context.Context, req SendEmailInput) (SendEmailOutput, error)
}

type SendEmailInput struct {
	ToEmail     string
	ToName      string
	Subject     string
	HTMLContent string
	TextContent string
}

type SendEmailOutput struct {
	MessageID string
}

type MailjetProvider struct {
	apiKey      string
	apiSecret   string
	senderEmail string
	senderName  string
	httpClient  *http.Client
}

func NewMailjetProvider(cfg config.MailjetConfig) *MailjetProvider {
	return &MailjetProvider{
		apiKey:      cfg.APIKey,
		apiSecret:   cfg.APISecret,
		senderEmail: cfg.SenderEmail,
		senderName:  cfg.SenderName,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type mailjetRequestBody struct {
	Messages []mailjetMessage `json:"Messages"`
}

type mailjetMessage struct {
	From     mailjetContact   `json:"From"`
	To       []mailjetContact `json:"To"`
	Subject  string           `json:"Subject"`
	TextPart string           `json:"TextPart,omitempty"`
	HTMLPart string           `json:"HTMLPart,omitempty"`
}

type mailjetContact struct {
	Email string `json:"Email"`
	Name  string `json:"Name,omitempty"`
}

type mailjetResponseBody struct {
	Messages []struct {
		Status string `json:"Status"`
		To     []struct {
			Email       string `json:"Email"`
			MessageUUID string `json:"MessageUUID"`
		} `json:"To"`
		Errors []struct {
			ErrorMessage string `json:"ErrorMessage"`
		} `json:"Errors"`
	} `json:"Messages"`
}

func (p *MailjetProvider) SendEmail(ctx context.Context, in SendEmailInput) (SendEmailOutput, error) {
	if in.ToEmail == "" {
		return SendEmailOutput{}, fmt.Errorf("to_email wajib diisi")
	}
	if in.Subject == "" {
		return SendEmailOutput{}, fmt.Errorf("subject wajib diisi")
	}
	if in.HTMLContent == "" && in.TextContent == "" {
		return SendEmailOutput{}, fmt.Errorf("html_content atau text_content wajib salah satu diisi")
	}

	body := mailjetRequestBody{
		Messages: []mailjetMessage{
			{
				From:     mailjetContact{Email: p.senderEmail, Name: p.senderName},
				To:       []mailjetContact{{Email: in.ToEmail, Name: in.ToName}},
				Subject:  in.Subject,
				TextPart: in.TextContent,
				HTMLPart: in.HTMLContent,
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return SendEmailOutput{}, fmt.Errorf("marshal payload mailjet: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, mailjetSendURL, bytes.NewReader(payload))
	if err != nil {
		return SendEmailOutput{}, fmt.Errorf("build request mailjet: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(p.apiKey, p.apiSecret)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return SendEmailOutput{}, fmt.Errorf("hubungi mailjet: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return SendEmailOutput{}, fmt.Errorf("baca response mailjet: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return SendEmailOutput{}, fmt.Errorf("mailjet status %d: %s", resp.StatusCode, string(rawBody))
	}

	var parsed mailjetResponseBody
	if err := json.Unmarshal(rawBody, &parsed); err != nil {
		return SendEmailOutput{}, fmt.Errorf("parse response mailjet: %w", err)
	}
	if len(parsed.Messages) == 0 {
		return SendEmailOutput{}, fmt.Errorf("response mailjet kosong")
	}

	msg := parsed.Messages[0]
	if msg.Status != "success" {
		errMsg := "unknown error"
		if len(msg.Errors) > 0 {
			errMsg = msg.Errors[0].ErrorMessage
		}
		return SendEmailOutput{}, fmt.Errorf("mailjet gagal kirim: %s", errMsg)
	}

	messageID := ""
	if len(msg.To) > 0 {
		messageID = msg.To[0].MessageUUID
	}

	return SendEmailOutput{MessageID: messageID}, nil
}
