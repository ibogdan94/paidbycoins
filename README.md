# Basic implementation of "Paid By Coins" API written by standard Go v1.9 library

## Example of usage

```go
package main

import (
	"os"
	"io/ioutil"
	"encoding/json"
	pbc "github.com/ibogdan94/paidbycoins"
	"fmt"
)

var PaidByCoinsClient pbc.PaidByCoins

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
	//resp, err := PaidByCoinsClient.CreatePayment(pbc.Invoice{
	//	"BTC",
	//	"AUD",
	//	234.0,
	//	"test description",
	//	pbc.CustomerDetails{
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

	body, err := ioutil.ReadAll(resp.Body)


	fmt.Println(string(body))
}
```