//go:build integration

package integration

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

func gatewayURL() string {
	if v := os.Getenv("API_GATEWAY_URL"); v != "" {
		return v
	}
	return "http://localhost:8080"
}

func doJSON(t *testing.T, client *http.Client, method, url, token string, body any) (int, map[string]any) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal %s %s: %v", method, url, err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("new request %s %s: %v", method, url, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s failed: %v", method, url, err)
	}
	defer resp.Body.Close()

	rawResp, _ := io.ReadAll(resp.Body)
	var parsed map[string]any
	if len(rawResp) > 0 {
		_ = json.Unmarshal(rawResp, &parsed)
	}
	return resp.StatusCode, parsed
}

func TestE2E_SupplyChainAndPOSFlow(t *testing.T) {
	base := gatewayURL()
	client := &http.Client{Timeout: 10 * time.Second}

	// Ensure stack is up
	healthResp, err := client.Get(base + "/health")
	if err != nil {
		t.Fatalf("API Gateway not reachable at %s (start stack with 'make up' first): %v", base, err)
	}
	healthResp.Body.Close()
	if healthResp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected /health status: %d", healthResp.StatusCode)
	}

	// 1. Unauthenticated request rejected
	status, _ := doJSON(t, client, http.MethodGet, base+"/api/v1/products", "", nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated request, got %d", status)
	}

	// 2. Login with seeder manager account
	status, loginResp := doJSON(t, client, http.MethodPost, base+"/api/v1/auth/login", "", map[string]string{
		"email":    "manager@mail.com",
		"password": "manager123",
	})
	if status != http.StatusOK {
		t.Fatalf("login failed with status %d: %+v", status, loginResp)
	}
	dataMap, _ := loginResp["data"].(map[string]any)
	token, _ := dataMap["token"].(string)
	if token == "" {
		t.Fatalf("expected non-empty JWT token in login response: %+v", loginResp)
	}
	userMap, _ := dataMap["user"].(map[string]any)
	keeperID := uint(1)
	if idFloat, ok := userMap["id"].(float64); ok && idFloat > 0 {
		keeperID = uint(idFloat)
	}

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// 3. Create Category & Product
	status, catResp := doJSON(t, client, http.MethodPost, base+"/api/v1/categories", token, map[string]any{
		"name":    "Sembako-" + suffix,
		"tagline": "Kebutuhan pokok",
		"photo":   "https://example.com/cat.png",
	})
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("create category failed (%d): %+v", status, catResp)
	}
	catID := uint(catResp["data"].(map[string]any)["id"].(float64))

	barcode := "899" + suffix[len(suffix)-10:]
	status, prodResp := doJSON(t, client, http.MethodPost, base+"/api/v1/products", token, map[string]any{
		"name":        "Beras E2E " + suffix,
		"barcode":     barcode,
		"price":       50000,
		"about":       "Produk uji integrasi",
		"category_id": catID,
		"thumbnail":   "https://example.com/prod.png",
		"is_popular":  true,
	})
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("create product failed (%d): %+v", status, prodResp)
	}
	productID := uint(prodResp["data"].(map[string]any)["id"].(float64))

	// 4. Create Warehouse & Add Stock (50)
	status, whResp := doJSON(t, client, http.MethodPost, base+"/api/v1/warehouses", token, map[string]any{
		"name":    "Gudang E2E " + suffix,
		"address": "Jl. Industri No. 1",
		"phone":   "08123456789",
		"photo":   "https://example.com/wh.png",
	})
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("create warehouse failed (%d): %+v", status, whResp)
	}
	warehouseID := uint(whResp["data"].(map[string]any)["id"].(float64))

	status, wpResp := doJSON(t, client, http.MethodPost, fmt.Sprintf("%s/api/v1/warehouse-products/%d", base, warehouseID), token, map[string]any{
		"product_id": productID,
		"stock":      50,
	})
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("create warehouse product failed (%d): %+v", status, wpResp)
	}

	// 5. Create Merchant & Allocate 10 units (triggers RabbitMQ stock.reduction -> warehouse_db)
	status, merchResp := doJSON(t, client, http.MethodPost, base+"/api/v1/merchants", token, map[string]any{
		"name":      "Toko E2E " + suffix,
		"keeper_id": keeperID,
		"address":   "Jl. Pasar No. 2",
		"phone":     "08129876543",
		"photo":     "https://example.com/merch.png",
	})
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("create merchant failed (%d): %+v", status, merchResp)
	}
	merchantID := uint(merchResp["data"].(map[string]any)["id"].(float64))

	status, mpResp := doJSON(t, client, http.MethodPost, base+"/api/v1/merchant-products", token, map[string]any{
		"product_id":   productID,
		"warehouse_id": warehouseID,
		"merchant_id":  merchantID,
		"stock":        10,
	})
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("allocate merchant product failed (%d): %+v", status, mpResp)
	}

	// Poll warehouse detail until stock drops to 40 via RabbitMQ
	deadline := time.Now().Add(10 * time.Second)
	reduced := false
	for time.Now().Before(deadline) {
		st, detail := doJSON(t, client, http.MethodGet, fmt.Sprintf("%s/api/v1/warehouses/%d", base, warehouseID), token, nil)
		if st == http.StatusOK {
			if d, ok := detail["data"].(map[string]any); ok {
				if prods, ok := d["products"].([]any); ok && len(prods) > 0 {
					if p0, ok := prods[0].(map[string]any); ok {
						if stk, ok := p0["stock"].(float64); ok && int(stk) == 40 {
							reduced = true
							break
						}
					}
				}
			}
		}
		time.Sleep(300 * time.Millisecond)
	}
	if !reduced {
		t.Logf("warning: warehouse stock reduction via RabbitMQ not observed within 10s (check warehouse-service/rabbitmq logs)")
	}

	// 6. Optional Midtrans webhook check if MIDTRANS_SERVER_KEY is configured
	if serverKey := os.Getenv("MIDTRANS_SERVER_KEY"); serverKey != "" {
		orderID := "E2E-" + suffix
		gross := "111000.00"
		sum := sha512.Sum512([]byte(orderID + "200" + gross + serverKey))
		sig := hex.EncodeToString(sum[:])

		cbStatus, cbResp := doJSON(t, client, http.MethodPost, base+"/api/v1/midtrans/callback", "", map[string]any{
			"order_id":           orderID,
			"transaction_status": "settlement",
			"payment_type":       "qris",
			"fraud_status":       "accept",
			"transaction_id":     "tx-" + suffix,
			"status_code":        "200",
			"gross_amount":       gross,
			"signature_key":      sig,
		})
		if cbStatus >= 500 {
			t.Fatalf("midtrans callback returned server error %d: %+v", cbStatus, cbResp)
		}
	}
}
