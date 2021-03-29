package mpesa

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	PRODUCTION             = "https://api.safaricom.co.ke/"
	SANDBOX                = "https://sandbox.safaricom.co.ke/"
	CONSUMER_KEY           = "1WyPp7K9UoSABpf0JaXuQUYOX7V4xC3T"
	CONSUMER_SECRET        = "0CGmzKJAHZJtWuGS"
	PAYBILL                = "174379"
	PASS_KEY               = "bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919"
	CustomerPayBillOnline  = "CustomerPayBillOnline"
	CustomerBuyGoodsOnline = "CustomerBuyGoodsOnline"
	STK_CALLBACK           = "https://peternjeru.co.ke/safdaraja/api/callback.php"
)

type Oauth struct {
	Token  string `json:"access_token"`
	Expire string `json:"expires_in"`
}

func token() string {
	cred := base64.StdEncoding.Strict().EncodeToString([]byte(CONSUMER_KEY + ":" + CONSUMER_SECRET))
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, SANDBOX+"oauth/v1/generate?grant_type=client_credentials", nil)

	if err != nil {
		fmt.Println(err.Error())
	}
	req.Header.Add("Authorization", "Basic "+cred)
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	var responseObject Oauth

	json.Unmarshal(bodyBytes, &responseObject)
	tok := responseObject.Token

	return tok
}

func stk(phone string) string {
	client := &http.Client{}

	t := time.Now()
	tf := t.Format("20060102150405")
	cred := base64.StdEncoding.Strict().EncodeToString([]byte(PAYBILL + PASS_KEY + tf))

	jsonData := map[string]string{
		"BusinessShortCode": PAYBILL,
		"Password":          cred,
		"Timestamp":         tf,
		"TransactionType":   CustomerPayBillOnline,
		"Amount":            "1",
		"PartyA":            phone,
		"PartyB":            "174379",
		"PhoneNumber":       phone,
		"CallBackURL":       STK_CALLBACK,
		"AccountReference":  "account",
		"TransactionDesc":   "test",
	}
	jsonValue, _ := json.Marshal(jsonData)
	req, err := http.NewRequest(http.MethodPost, SANDBOX+"mpesa/stkpush/v1/processrequest", bytes.NewBuffer(jsonValue))

	if err != nil {
		fmt.Println(err)
	}

	req.Header.Add("Authorization", "Bearer "+token())
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	byteData, _ := ioutil.ReadAll(resp.Body)

	return string(byteData)
}

func C2BRegister() string {
	client := &http.Client{}

	jsonData := map[string]string{
		"ShortCode":       "601426",
		"ResponseType":    "Completed",
		"ConfirmationURL": "https://saas1.apartmentaly.com/confirmation",
		"ValidationURL":   "https://saas1.apartmentaly.com/validate",
	}
	jsonValue, _ := json.Marshal(jsonData)

	req, err := http.NewRequest(http.MethodPost, SANDBOX+"mpesa/c2b/v1/registerurl", bytes.NewBuffer(jsonValue))

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Authorization", "Bearer "+token())
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	byteData, _ := ioutil.ReadAll(resp.Body)

	return string(byteData)
}

func C2BRequest() {
	client := &http.Client{}

	jsonData := map[string]string{
		"ShortCode":     "601426",
		"CommandID":     CustomerPayBillOnline,
		"Amount":        "1",
		"Msisdn":        "254708374149",
		"BillRefNumber": "account",
	}
	jsonValue, _ := json.Marshal(jsonData)

	req, err := http.NewRequest(http.MethodPost, SANDBOX+"mpesa/c2b/v1/simulate", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(token())
	req.Header.Add("Authorization", "Bearer "+token())
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
	}
	byteValue, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(byteValue))
}

func C2BCallback(r http.ResponseWriter, req *http.Request) {
	fmt.Println("called")
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
	jsonData := map[string]interface{}{
		"ResultCode": 0,
		"ResultDesc": "Success",
	}
	jsonVaue, _ := json.Marshal(jsonData)
	r.Write(jsonVaue)
}

func pull_Register() string {
	client := &http.Client{}
	jsonData := map[string]string{
		"ShortCode":       "600000",
		"RequestType":     "Pull",
		"NominatedNumber": "0722000000",
		"CallBackURL":     "https://peternjeru.co.ke/safdaraja/api/callback.php",
	}
	jsonValue, _ := json.Marshal(jsonData)

	req, err := http.NewRequest(http.MethodPost, SANDBOX+"pulltransactions/v1/register", bytes.NewBuffer(jsonValue))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+token())
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	byteData, _ := ioutil.ReadAll(resp.Body)
	return string(byteData)
}

func Pull_Transaction() string {
	client := &http.Client{}
	jsonData := map[string]string{
		"ShortCode":   "600000",
		"StartDate":   "2020-08-04 8:36:00",
		"EndDate":     "2020-08-16 10:10:000",
		"OffSetValue": "0",
	}
	jsonValue, _ := json.Marshal(jsonData)

	req, err := http.NewRequest(http.MethodPost, SANDBOX+"pulltransactions/v1/query", bytes.NewBuffer(jsonValue))

	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+token())
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	byteValue, _ := ioutil.ReadAll(resp.Body)

	return string(byteValue)
}
