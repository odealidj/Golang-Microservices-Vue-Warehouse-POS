package rabbitmq

import (
	"encoding/json"
	"testing"
	"time"
)

func TestStockReductionEventContract(t *testing.T) {
	if ExchangeName != "warehouse_events" {
		t.Fatalf("expected ExchangeName warehouse_events, got %s", ExchangeName)
	}
	if RoutingKey != "stock.reduction" {
		t.Fatalf("expected RoutingKey stock.reduction, got %s", RoutingKey)
	}

	event := StockReductionEvent{
		WarehouseID: 3,
		ProductID:   12,
		Stock:       25,
		MerchantID:  7,
		Timestamp:   time.Unix(1700000000, 0).UTC(),
	}

	raw, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if _, ok := decoded["warehouse_id"]; !ok {
		t.Fatalf("expected JSON key 'warehouse_id', got payload: %s", string(raw))
	}
	if _, typo := decoded["warhouse_id"]; typo {
		t.Fatalf("unexpected typo key 'warhouse_id' in payload: %s", string(raw))
	}
}
