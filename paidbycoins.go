package paidbycoins

import (
	b64 "encoding/base64"
	"fmt"
	"time"
	"crypto/sha256"
	"crypto/hmac"
	"net/http"
	"crypto/md5"
	"strings"
	"encoding/json"
	"io/ioutil"
)

type Invoice struct {
	CryptoCurrency string //"BTC", "LTC", "ETH"
	Currency       string //"AUD"
	Amount         float64
	Description    string
	Detail         CustomerDetails
}

type CustomerDetails struct {
	Email         string
	MerchantRefNo string
	FullName      string
	FirstName     string
	LastName      string
	ContactNo     string
	BirthDate     string
	Address       string
	City          string
	State         string
	Zip           string
	Country       string
}

type PaidByCoinsApiCaller interface {
	GenerateNonce() int64
	GetRates() (*Response, error)
	GetPaymentStatus(paymentId int) (*Response, error)
	CreatePayment(invoice Invoice) (*Response, error)
}

type Response struct {
	StatusCode int
	Status string
	Message *string
	Data interface{}
}

type PaidByCoins struct {
	BaseURL string `json:"api_endpoint"`
	MID     string `json:"merchant_id"`
	APIKey  string `json:"api_key"`
}

func (p PaidByCoins) GenerateNonce() int64 {
	return time.Now().UnixNano()
}

func (p PaidByCoins) GetRates() (*Response, error) {
	return makeApiRequest(p, "GET", "/v1/cli/rates", "")
}

func (p PaidByCoins) GetPaymentStatus(paymentId int) (*Response, error) {
	return makeApiRequest(p, "GET", fmt.Sprintf("/v1/cli/status/%d", paymentId), "")
}

func (p PaidByCoins) CreatePayment(invoice Invoice) (*Response, error) {
	jsonBytes, err := json.Marshal(invoice)

	if err != nil {
		return nil, err
	}

	return makeApiRequest(p, "POST", "/v1/cli/createpayment", string(jsonBytes))
}

func makeApiRequest(p PaidByCoins, method string, endpoint string, payload string) (*Response, error) {
	nonce := p.GenerateNonce()
	now := time.Now()
	timestamp := now.Format("20060102 15:04:05")

	var base64Payload string

	if payload != "" {
		base64Payload = b64.StdEncoding.EncodeToString(computeMD5Hash(payload))
	}

	signatureRawData := fmt.Sprintf(
		"%s%s%s%s%d%s",
		p.MID,
		strings.ToUpper(method),
		p.BaseURL+endpoint,
		timestamp,
		nonce,
		base64Payload,
	)

	requestSignatureBase64String := b64.StdEncoding.EncodeToString(computeHmac256(p.APIKey, signatureRawData))

	sign := fmt.Sprintf("%s||%s||%d||%s", p.MID, requestSignatureBase64String, nonce, timestamp)

	req, err := http.NewRequest(method, p.BaseURL+endpoint, strings.NewReader(payload))

	if err != nil {
		return nil, err
	}

	req.Header.Add("pbcx", sign)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var srvResp Response

	err = json.Unmarshal(body, &srvResp)

	if err != nil {
		return nil, err
	}

	srvResp.StatusCode = resp.StatusCode

	return &srvResp, nil
}

func computeHmac256(secret string, payload string) []byte {
	key, _ := b64.StdEncoding.DecodeString(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(payload))

	return h.Sum(nil)
}

func computeMD5Hash(payload string) []byte {
	h := md5.New()
	h.Write([]byte(payload))

	return h.Sum(nil)
}