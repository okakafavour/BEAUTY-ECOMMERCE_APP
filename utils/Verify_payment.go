package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// PaystackVerifyResponse represents Paystack's verification response
type PaystackVerifyResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Status    string `json:"status"` // "success", "failed", "abandoned"
		Amount    int    `json:"amount"`
		Reference string `json:"reference"`
		Email     string `json:"customer_email"`
	} `json:"data"`
}

// VerifyPaystackPayment checks if a payment was successful
func VerifyPaystackPayment(reference string) (*PaystackVerifyResponse, error) {
	secret := os.Getenv("PAYSTACK_SECRET_KEY")
	baseURL := os.Getenv("PAYSTACK_BASE_URL")

	if secret == "" || baseURL == "" {
		return nil, fmt.Errorf("paystack secret or base URL not set")
	}

	url := fmt.Sprintf("%s/transaction/verify/%s", baseURL, reference)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+secret)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)
	fmt.Println("🔹 Paystack verification response:", string(bodyBytes)) // <-- log full response

	var payRes PaystackVerifyResponse
	if err := json.Unmarshal(bodyBytes, &payRes); err != nil {
		return nil, err
	}

	if !payRes.Status || payRes.Data.Status != "success" {
		return nil, fmt.Errorf("payment not successful")
	}

	return &payRes, nil
}
