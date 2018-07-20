package main

import (
	b64 "encoding/base64"
	"fmt"
	"time"
	"crypto/sha256"
	"crypto/hmac"
	"net/http"
	"io/ioutil"
	"crypto/md5"
	"strings"
	"encoding/json"
	"os"
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

type PaidByCoinsApiMethods interface {
	GetRates() ([]byte, error)
	GetPaymentStatus(paymentId int) ([]byte, error)
	CreatePayment(invoice Invoice) ([]byte, error)
}

type PaidByCoins struct {
	BaseURL string `json:"api_endpoint"`
	MID     string `json:"merchant_id"`
	APIKey  string `json:"api_key"`
}

func (p PaidByCoins) GetRates() ([]byte, error) {
	bytes, err := makeApiRequest(p, "GET", "/v1/cli/rates", "")

	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (p PaidByCoins) GetPaymentStatus(paymentId int) ([]byte, error) {
	bytes, err := makeApiRequest(p, "GET", fmt.Sprintf("/v1/cli/status/%d", paymentId), "")

	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (p PaidByCoins) CreatePayment(invoice Invoice) ([]byte, error) {
	jsonBytes, err := json.Marshal(invoice)

	if err != nil {
		return nil, err
	}

	bytes, err := makeApiRequest(p, "POST", "/v1/cli/createpayment", string(jsonBytes))

	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func makeApiRequest(p PaidByCoins, method string, endpoint string, payload string) ([]byte, error) {
	now := time.Now()
	Nonce := now.UnixNano()
	timestamp := now.Format("20060102 15:04:05")

	var base64Payload string

	if payload != "" {
		base64Payload = computeMD5Hash(payload)
	}

	signatureRawData := fmt.Sprintf(
		"%s%s%s%s%d%s",
		p.MID,
		strings.ToUpper(method),
		p.BaseURL+endpoint,
		timestamp,
		Nonce,
		base64Payload,
	)

	//fmt.Printf("signatureRawData: %s \n", signatureRawData)

	requestSignatureBase64String := computeHmac256(p.APIKey, signatureRawData)

	sign := fmt.Sprintf("%s||%s||%d||%s", p.MID, requestSignatureBase64String, Nonce, timestamp)

	//fmt.Printf("sign: %s \n", sign)

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

	return body, nil
}

func computeHmac256(secret string, payload string) string {
	key, _ := b64.StdEncoding.DecodeString(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(payload))

	return b64.StdEncoding.EncodeToString(h.Sum(nil))
}

func computeMD5Hash(payload string) string {
	h := md5.New()
	h.Write([]byte(payload))

	return b64.StdEncoding.EncodeToString(h.Sum(nil))
}

var PaidByCoinsClient PaidByCoins

func parseJSONConfig() {
	pwd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	payload, err := ioutil.ReadFile(pwd + "/config.json")

	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(payload, &PaidByCoinsClient); err != nil {
		panic(err)
	}
}

func init() {
	parseJSONConfig()
}

func main() {
	//resp, err := PaidByCoinsClient.CreatePayment(Invoice{
	//	"BTC",
	//	"AUD",
	//	234.0,
	//	"test description",
	//	CustomerDetails{
	//		"user@gmail.com",
	//		"9876543210",
	//		"Full Name",
	//		"Firstname",
	//		"Last Name",
	//		"9876543210",
	//		"1993-01-26",
	//		"test Address",
	//		"texas",
	//		"sydney",
	//		"64100",
	//		"Australia",
	//	},
	//})

	//resp, err := PaidByCoinsClient.GetPaymentStatus(201807200266)

	resp, err := PaidByCoinsClient.GetRates()

	if err != nil {
		panic(err)
	}

	fmt.Println(string(resp))
}
