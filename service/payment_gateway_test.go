package service

import (
	"net/url"
	"testing"

	"github.com/QuantumNous/new-api/setting/operation_setting"
)

func TestDetectPaymentGatewayType(t *testing.T) {
	original := operation_setting.PayAddress
	t.Cleanup(func() {
		operation_setting.PayAddress = original
	})

	operation_setting.PayAddress = "https://api.xunhupay.com/payment/do.html"
	if DetectPaymentGatewayType() != GatewayTypeXunhuPay {
		t.Fatalf("expected xunhupay gateway detection")
	}

	operation_setting.PayAddress = "https://pay.example.com"
	if DetectPaymentGatewayType() != GatewayTypeEpay {
		t.Fatalf("expected epay gateway detection")
	}
}

func TestSignAndVerifyXunhuParams(t *testing.T) {
	params := map[string]string{
		"appid":          "201906177554",
		"trade_order_id": "USR3NOOfbKKV1773330755",
		"total_fee":      "7.30",
		"title":          "TUC1",
		"time":           "1773330755",
		"notify_url":     "https://api.otcbot.com/api/user/epay/notify",
		"nonce_str":      "nonce1234",
	}

	signature := signXunhuParams(params, "secret")
	if signature == "" {
		t.Fatalf("expected non-empty signature")
	}

	params["hash"] = signature
	if !verifyXunhuSignature(params, "secret") {
		t.Fatalf("expected signature verification to pass")
	}

	params["status"] = "OD"
	if verifyXunhuSignature(params, "wrong-secret") {
		t.Fatalf("expected signature verification to fail with wrong secret")
	}
}

func TestGatewaySiteName(t *testing.T) {
	parsed, err := url.Parse("https://api.otcbot.com/console/log")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if got := normalizeGatewayHost(parsed); got != "https://api.otcbot.com" {
		t.Fatalf("unexpected host normalization result: %s", got)
	}
	if got := gatewaySiteName(parsed); got != "api.otcbot.com" {
		t.Fatalf("unexpected site name result: %s", got)
	}
}
