package midtrans

import (
	"crypto/sha512"
	"encoding/hex"
	"micro-warehouse/transaction-service/configs"
	"testing"
)

func computeTestSignature(orderID, statusCode, grossAmount, serverKey string) string {
	sum := sha512.Sum512([]byte(orderID + statusCode + grossAmount + serverKey))
	return hex.EncodeToString(sum[:])
}

func TestVerifySignature_Valid(t *testing.T) {
	svc := NewMidtransService(&configs.Config{
		Midtrans: configs.Midtrans{ServerKey: "SB-Mid-server-TESTKEY"},
	})

	sig := computeTestSignature("ORDER_123_1", "200", "111000.00", "SB-Mid-server-TESTKEY")
	if !svc.VerifySignature("ORDER_123_1", "200", "111000.00", sig) {
		t.Fatal("expected valid Midtrans signature to pass verification")
	}
}

func TestVerifySignature_TamperedAmount(t *testing.T) {
	svc := NewMidtransService(&configs.Config{
		Midtrans: configs.Midtrans{ServerKey: "SB-Mid-server-TESTKEY"},
	})

	sig := computeTestSignature("ORDER_123_1", "200", "111000.00", "SB-Mid-server-TESTKEY")
	if svc.VerifySignature("ORDER_123_1", "200", "1000.00", sig) {
		t.Fatal("expected tampered gross_amount to fail signature verification")
	}
}

func TestVerifySignature_MissingServerKey(t *testing.T) {
	svc := NewMidtransService(&configs.Config{})
	if svc.VerifySignature("ORDER_123_1", "200", "111000.00", "deadbeef") {
		t.Fatal("expected empty server key to fail closed")
	}
}
