package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Sender interface {
	SendMagicLink(to, verifyURL string) error
}

type resendSender struct {
	apiKey string
	from   string
	client *http.Client
}

func NewResendSender() Sender {
	return &resendSender{
		apiKey: os.Getenv("RESEND_API_KEY"),
		from:   os.Getenv("RESEND_FROM"),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (s *resendSender) SendMagicLink(to, verifyURL string) error {
	payload := resendPayload{
		From:    s.from,
		To:      []string{to},
		Subject: "Your Formful login link",
		HTML:    fmt.Sprintf(`<p><a href="%s">Click here to log in</a> (expires in 15 minutes)</p>`, verifyURL),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend: status %d", resp.StatusCode)
	}
	return nil
}
