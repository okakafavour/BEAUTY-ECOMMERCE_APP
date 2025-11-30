package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
)

type PaystackInitializeRequest struct {
	Email     string `json:"email"`
	Amount    int    `json:"amount"`
	Reference string `json:"reference"`
}

type PaystackInitializeResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

func PaystackInitialize(email string, amount float64, reference string) (*PaystackInitializeResponse, error) {
	// Read env variables inside function (ensures they are loaded)
	paystackSecret := os.Getenv("PAYSTACK_SECRET_KEY")
	paystackBase := os.Getenv("PAYSTACK_BASE_URL")
	callbackURL := os.Getenv("PAYSTACK_CALLBACK_URL")

	if paystackSecret == "" || paystackBase == "" {
		return nil, errors.New("paystack secret or base URL not set")
	}

	if callbackURL == "" {
		return nil, errors.New("callback URL not set")
	}

	// Convert Naira to Kobo
	koboAmount := int(amount * 100)

	// Build request payload
	payload := map[string]interface{}{
		"email":        email,
		"amount":       koboAmount,
		"reference":    reference,
		"callback_url": callbackURL,
	}

	body, _ := json.Marshal(payload)

	// Full URL
	url := paystackBase + "/transaction/initialize"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Set headers correctly
	req.Header.Set("Authorization", "Bearer "+paystackSecret)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var payRes PaystackInitializeResponse
	if err := json.NewDecoder(res.Body).Decode(&payRes); err != nil {
		return nil, err
	}

	if !payRes.Status {
		return nil, errors.New(payRes.Message)
	}

	return &payRes, nil
}
