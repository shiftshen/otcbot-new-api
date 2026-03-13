package service

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/Calcium-Ion/go-epay/epay"
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
)

const (
	xunhuVersion          = "1.1"
	xunhuSuccessStatus    = "OD"
	paymentPluginIdentity = "new-api"
)

type GatewayType string

const (
	GatewayTypeEpay     GatewayType = "epay"
	GatewayTypeXunhuPay GatewayType = "xunhupay"
)

type PaymentPurchaseArgs struct {
	PaymentMethod string
	TradeNo       string
	Title         string
	Money         string
	NotifyURL     *url.URL
	ReturnURL     *url.URL
}

type PaymentPurchaseResult struct {
	URL     string
	Params  map[string]string
	PayLink string
	Raw     map[string]string
}

type PaymentVerifyResult struct {
	TradeNo      string
	TradeStatus  string
	VerifyStatus bool
	Raw          map[string]string
}

func DetectPaymentGatewayType() GatewayType {
	payAddress := strings.ToLower(operation_setting.PayAddress)
	if strings.Contains(payAddress, "xunhupay.com/payment/do.html") ||
		strings.Contains(payAddress, "dpweixin.com/payment/do.html") {
		return GatewayTypeXunhuPay
	}
	return GatewayTypeEpay
}

func HasOnlinePaymentConfig() bool {
	return operation_setting.PayAddress != "" &&
		operation_setting.EpayId != "" &&
		operation_setting.EpayKey != ""
}

func GetPaymentMethodAvailability(method string) (bool, string) {
	switch DetectPaymentGatewayType() {
	case GatewayTypeXunhuPay:
		if method == "wxpay" {
			return false, "当前仅开通支付宝支付"
		}
	}
	return true, ""
}

func CreatePayment(args *PaymentPurchaseArgs) (*PaymentPurchaseResult, error) {
	if enabled, reason := GetPaymentMethodAvailability(args.PaymentMethod); !enabled {
		return nil, errors.New(reason)
	}
	switch DetectPaymentGatewayType() {
	case GatewayTypeXunhuPay:
		return createXunhuPayment(args)
	default:
		return createEpayPayment(args)
	}
}

func VerifyPayment(params map[string]string) (*PaymentVerifyResult, error) {
	switch DetectPaymentGatewayType() {
	case GatewayTypeXunhuPay:
		return verifyXunhuPayment(params)
	default:
		return verifyEpayPayment(params)
	}
}

func GetEpayClient() *epay.Client {
	if !HasOnlinePaymentConfig() {
		return nil
	}
	withURL, err := epay.NewClient(&epay.Config{
		PartnerID: operation_setting.EpayId,
		Key:       operation_setting.EpayKey,
	}, operation_setting.PayAddress)
	if err != nil {
		return nil
	}
	return withURL
}

func createEpayPayment(args *PaymentPurchaseArgs) (*PaymentPurchaseResult, error) {
	client := GetEpayClient()
	if client == nil {
		return nil, fmt.Errorf("payment gateway is not configured")
	}

	uri, params, err := client.Purchase(&epay.PurchaseArgs{
		Type:           args.PaymentMethod,
		ServiceTradeNo: args.TradeNo,
		Name:           args.Title,
		Money:          args.Money,
		Device:         epay.PC,
		NotifyUrl:      args.NotifyURL,
		ReturnUrl:      args.ReturnURL,
	})
	if err != nil {
		return nil, err
	}

	return &PaymentPurchaseResult{
		URL:    uri,
		Params: params,
	}, nil
}

func verifyEpayPayment(params map[string]string) (*PaymentVerifyResult, error) {
	client := GetEpayClient()
	if client == nil {
		return nil, fmt.Errorf("payment gateway is not configured")
	}

	verifyInfo, err := client.Verify(params)
	if err != nil {
		return nil, err
	}

	return &PaymentVerifyResult{
		TradeNo:      verifyInfo.ServiceTradeNo,
		TradeStatus:  verifyInfo.TradeStatus,
		VerifyStatus: verifyInfo.VerifyStatus,
		Raw:          cloneStringMap(params),
	}, nil
}

type xunhuCreatePaymentResponse struct {
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	URL       string `json:"url"`
	URLQrcode string `json:"url_qrcode"`
	OpenID    any    `json:"openid"`
	Hash      string `json:"hash"`
}

func createXunhuPayment(args *PaymentPurchaseArgs) (*PaymentPurchaseResult, error) {
	params := map[string]string{
		"version":        xunhuVersion,
		"appid":          operation_setting.EpayId,
		"trade_order_id": args.TradeNo,
		"total_fee":      args.Money,
		"title":          args.Title,
		"time":           fmt.Sprintf("%d", time.Now().Unix()),
		"notify_url":     args.NotifyURL.String(),
		"return_url":     args.ReturnURL.String(),
		"plugins":        paymentPluginIdentity,
		"nonce_str":      common.GetRandomString(16),
	}

	if args.PaymentMethod == "wxpay" {
		params["type"] = "WAP"
		params["wap_url"] = normalizeGatewayHost(args.ReturnURL)
		params["wap_name"] = gatewaySiteName(args.ReturnURL)
	}

	params["hash"] = signXunhuParams(params, operation_setting.EpayKey)

	body := url.Values{}
	for key, value := range params {
		body.Set(key, value)
	}

	req, err := http.NewRequest(http.MethodPost, operation_setting.PayAddress, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result xunhuCreatePaymentResponse
	if err = json.Unmarshal(payload, &result); err != nil {
		return nil, fmt.Errorf("xunhupay returned invalid response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("xunhupay request failed: http %d", resp.StatusCode)
	}

	rawResult := map[string]interface{}{}
	if err = json.Unmarshal(payload, &rawResult); err != nil {
		return nil, fmt.Errorf("xunhupay returned invalid response: %w", err)
	}
	rawMap := stringifyMap(rawResult)
	if result.Hash != "" && !verifyXunhuSignature(rawMap, operation_setting.EpayKey) {
		return nil, fmt.Errorf("xunhupay response signature verification failed")
	}

	if result.ErrCode != 0 {
		if result.ErrMsg == "" {
			result.ErrMsg = "xunhupay returned an unknown error"
		}
		return nil, errors.New(result.ErrMsg)
	}
	if result.URL == "" {
		return nil, fmt.Errorf("xunhupay did not return a payment url")
	}

	return &PaymentPurchaseResult{
		PayLink: result.URL,
		Raw:     rawMap,
	}, nil
}

func verifyXunhuPayment(params map[string]string) (*PaymentVerifyResult, error) {
	if !verifyXunhuSignature(params, operation_setting.EpayKey) {
		return &PaymentVerifyResult{
			TradeNo:      params["trade_order_id"],
			TradeStatus:  params["status"],
			VerifyStatus: false,
			Raw:          cloneStringMap(params),
		}, nil
	}

	return &PaymentVerifyResult{
		TradeNo:      params["trade_order_id"],
		TradeStatus:  params["status"],
		VerifyStatus: true,
		Raw:          cloneStringMap(params),
	}, nil
}

func IsPaymentSuccessStatus(status string) bool {
	switch DetectPaymentGatewayType() {
	case GatewayTypeXunhuPay:
		return status == xunhuSuccessStatus
	default:
		return status == epay.StatusTradeSuccess
	}
}

func signXunhuParams(params map[string]string, secret string) string {
	filtered := make([]string, 0, len(params))
	for key, value := range params {
		if key == "hash" || value == "" {
			continue
		}
		filtered = append(filtered, key)
	}
	sort.Strings(filtered)

	buf := bytes.NewBuffer(nil)
	for index, key := range filtered {
		if index > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(key)
		buf.WriteByte('=')
		buf.WriteString(params[key])
	}
	sum := md5.Sum(append(buf.Bytes(), []byte(secret)...))
	return hex.EncodeToString(sum[:])
}

func verifyXunhuSignature(params map[string]string, secret string) bool {
	if params["hash"] == "" {
		return false
	}
	return signXunhuParams(params, secret) == params["hash"]
}

func cloneStringMap(input map[string]string) map[string]string {
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func stringifyMap(input map[string]interface{}) map[string]string {
	output := make(map[string]string, len(input))
	for key, value := range input {
		switch typed := value.(type) {
		case string:
			output[key] = typed
		case float64:
			if typed == float64(int64(typed)) {
				output[key] = fmt.Sprintf("%d", int64(typed))
			} else {
				output[key] = strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f", typed), "0"), ".")
			}
		case bool:
			if typed {
				output[key] = "true"
			} else {
				output[key] = "false"
			}
		default:
			output[key] = fmt.Sprintf("%v", typed)
		}
	}
	return output
}

func normalizeGatewayHost(rawURL *url.URL) string {
	if rawURL == nil {
		return ""
	}
	return fmt.Sprintf("%s://%s", rawURL.Scheme, rawURL.Host)
}

func gatewaySiteName(rawURL *url.URL) string {
	if rawURL == nil || rawURL.Host == "" {
		return "new-api"
	}
	host := rawURL.Hostname()
	if host == "" {
		host = "new-api"
	}
	if len(host) > 32 {
		return host[:32]
	}
	return host
}
