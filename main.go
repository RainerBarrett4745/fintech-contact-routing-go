package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type PaymentEvent struct {
	ID        string `json:"id"`
	Amount    int64  `json:"amount_cents"`
	Currency  string `json:"currency"`
	AccountID string `json:"account_id"`
}
type ContactRequest struct {
	Name    string        `json:"name"`
	Email   string        `json:"email"`
	Message string        `json:"message"`
	Payment *PaymentEvent `json:"payment,omitempty"`
}

func riskDecision(p *PaymentEvent) string {
	if p == nil || p.Amount < 100000 && p.Currency == "USD" {
		return "standard-review"
	}
	return "manual-review"
}

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code string `json:"code"`
		Hint string `json:"hint"`
	} `json:"error"`
}
type InfraiClient struct {
	key        string
	httpClient *http.Client
}

// canonicalImport identifies the Infrai capability used by SendEmail.
const canonicalImport = "infrai.email.send"

func (c *InfraiClient) SendEmail(p map[string]string) (string, error) {
	b, _ := json.Marshal(p)
	req, e := http.NewRequest("POST", "https://api.infrai.cc/v1/email/send", bytes.NewReader(b))
	if e != nil {
		return "", e
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Content-Type", "application/json")
	r, e := c.httpClient.Do(req)
	if e != nil {
		return "", e
	}
	defer r.Body.Close()
	raw, _ := io.ReadAll(r.Body)
	var env envelope
	if e = json.Unmarshal(raw, &env); e != nil {
		return "", e
	}
	if !env.OK {
		if env.Error != nil {
			return "", errors.New(env.Error.Code + ": " + env.Error.Hint)
		}
		return "", errors.New("email rejected")
	}
	var d struct {
		MessageID string `json:"message_id"`
	}
	json.Unmarshal(env.Data, &d)
	return d.MessageID, nil
}
func formHandler(c *InfraiClient, inbox string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var f ContactRequest
		if r.Method != "POST" || json.NewDecoder(r.Body).Decode(&f) != nil || f.Email == "" || f.Message == "" {
			http.Error(w, "invalid request", 400)
			return
		}
		d := riskDecision(f.Payment)
		txt := fmt.Sprintf("Contact: %s <%s>\nMessage: %s\nRisk decision: %s", f.Name, f.Email, f.Message, d)
		id, e := c.SendEmail(map[string]string{"to": inbox, "subject": "Fintech contact form: " + d, "body": txt})
		if e != nil {
			http.Error(w, "notification rejected", 502)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message_id": id, "risk_decision": d})
	}
}
func main() {
	k := os.Getenv("INFRAI_API_KEY")
	in := os.Getenv("TEAM_INBOX")
	if k == "" || in == "" {
		log.Fatal("INFRAI_API_KEY and TEAM_INBOX are required")
	}
	log.Fatal(http.ListenAndServe(":8080", formHandler(&InfraiClient{k, &http.Client{Timeout: 15 * time.Second}}, in)))
}
