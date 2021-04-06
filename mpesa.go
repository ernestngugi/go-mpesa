package mpesa

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/ttacon/libphonenumber"
)

const (
	ContentType = "application/json"
)

type Credentials struct {
	Environment     string
	CONSUMER_KEY    string
	CONSUMER_SECRET string
	PASS_KEY        string
	PAYBILL         string
	CALLBACK_URL    string
	CONFIRM_URL     string
	VALIDATE_URL    string
}

type Oauth struct {
	Token  string `json:"access_token"`
	Expire string `json:"expires_in"`
}
type STK_Request struct {
	BusinessShortCode string
	Password          string
	Timestamp         string
	TransactionType   string
	Amount            string
	PartyA            string
	PartyB            string
	PhoneNumber       string
	CallBackURL       string
	AccountReference  string
	TransactionDesc   string
}

type C2B_reg struct {
	ShortCode       string
	ResponseType    string
	ConfirmationURL string
	ValidationURL   string
}

type C2B_Register struct {
	ShortCode       string
	ResponseType    string
	ConfirmationURL string
	ValidationURL   string
}

type C2B_Request struct {
	ShortCode     string
	CommandID     string
	Amount        string
	Msisdn        string
	BillRefNumber string
}

type Pull_Reg struct {
	ShortCode       string
	RequestType     string
	NominatedNumber string
	CallBackURL     string
}

type Pull_trans struct {
	ShortCode   string
	StartDate   string
	EndDate     string
	OffSetValue string
}

func (c *Credentials) Creds() (tok string, er error) {
	if c.CONSUMER_KEY == "" {
		er = fmt.Errorf("consumer key cannot be empty")
		return
	}
	if c.CONSUMER_SECRET == "" {
		er = fmt.Errorf("consumer secret cannot be empty")
		return
	}

	switch c.Environment {
	case "sandbox", "production":
		break
	default:
		er = fmt.Errorf("available are: sandbox or production")
		return
	}
	return
}

func (c *Credentials) Token() (oauth *Oauth, err error) {
	token := &Oauth{}
	var URI string
	switch c.Environment {
	case "production":
		URI = "https://api.safaricom.co.ke/"
	case "sandbox":
		URI = "https://sandbox.safaricom.co.ke/"
	default:
		return
	}
	cred := base64.StdEncoding.Strict().EncodeToString([]byte(c.CONSUMER_KEY + ":" + c.CONSUMER_SECRET))
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, URI+"oauth/v1/generate?grant_type=client_credentials", nil)

	if err != nil {
		return
	}
	req.Header.Add("Authorization", "Basic "+cred)
	resp, err := client.Do(req)

	if err != nil {
		return
	}
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(bodyBytes, token)

	return token, err
}

func (c *Credentials) Stk(s *STK_Request) (str string, err error) {
	var URI string
	token, _ := c.Token()
	client := &http.Client{}
	code := regexp.MustCompile("^[0-9]+$")
	if !code.MatchString(s.BusinessShortCode) {
		err = fmt.Errorf("invalid shortcode")
		return
	}

	num, _ := libphonenumber.Parse(s.PhoneNumber, "KE")
	natSigNumber := libphonenumber.GetNationalSignificantNumber(num)

	if len(natSigNumber) > 9 {
		err = fmt.Errorf("provided number is not in KE")
		fmt.Println(err)
		return
	}

	switch s.TransactionType {
	case "CustomerPayBillOnline", "CustomerBuyGoodsOnline":
		break
	default:
		err = fmt.Errorf("available options are: customerpaybillonline or customerbuygoodsonline")
		return
	}

	if s.CallBackURL == "" {
		err = fmt.Errorf("callback is empty")
		return
	}

	if s.Amount < strconv.Itoa(1) {
		err = fmt.Errorf("amount cannot be less than 1")
		return
	}
	switch c.Environment {
	case "production":
		URI = "https://api.safaricom.co.ke/"
	case "sandbox":
		URI = "https://sandbox.safaricom.co.ke/"
	default:
		return
	}

	jsonValue, _ := json.Marshal(s)

	req, err := http.NewRequest(http.MethodPost, URI+"mpesa/stkpush/v1/processrequest", bytes.NewBuffer(jsonValue))

	if err != nil {
		fmt.Println(err)
	}

	req.Header.Add("Authorization", "Bearer "+token.Token)
	req.Header.Add("Content-Type", ContentType)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	byteData, _ := ioutil.ReadAll(resp.Body)

	return string(byteData), err
}

func (c *Credentials) C2BRegister(c2b *C2B_reg) (strr string, err error) {
	var URI string
	token, _ := c.Token()
	switch c.Environment {
	case "production":
		URI = "https://api.safaricom.co.ke/"
	case "sandbox":
		URI = "https://sandbox.safaricom.co.ke/"
	default:
		return
	}
	client := &http.Client{}

	jsonValue, _ := json.Marshal(c2b)

	req, err := http.NewRequest(http.MethodPost, URI+"mpesa/c2b/v1/registerurl", bytes.NewBuffer(jsonValue))

	if err != nil {
		return
	}
	req.Header.Add("Authorization", "Bearer "+token.Token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	byteData, _ := ioutil.ReadAll(resp.Body)

	return string(byteData), err
}

func C2BCallback(r http.ResponseWriter, req *http.Request) {
	ips := req.Header.Get("X-FORWARDED-FOR")
	iprange := strings.Split("196.201.214.200,196.201.214.206,196.201.213.114,196.201.214.207,196.201.214.208,196.201.213.44,196.201.212.127,196.201.212.128,196.201.212.129,196.201.212.132,196.201.212.136,196.201.212.138", ",")
	for _, ip := range iprange {
		if ips != ip {
			break
		}
	}
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

func (c *Credentials) Pull_Register(p *Pull_Reg) (str string, err error) {
	var URI string
	token, _ := c.Token()
	switch c.Environment {
	case "production":
		URI = "https://api.safaricom.co.ke/"
	case "sandbox":
		URI = "https://sandbox.safaricom.co.ke/"
	default:
		return
	}
	client := &http.Client{}
	jsonValue, _ := json.Marshal(p)

	req, err := http.NewRequest(http.MethodPost, URI+"pulltransactions/v1/register", bytes.NewBuffer(jsonValue))
	if err != nil {
		return
	}
	req.Header.Add("Authorization", "Bearer "+token.Token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		return
	}
	defer resp.Body.Close()

	byteData, _ := ioutil.ReadAll(resp.Body)
	return string(byteData), err
}

func (c *Credentials) Pull_Transaction(p *Pull_trans) (str string, err error) {
	var URI string
	token, _ := c.Token()
	switch c.Environment {
	case "production":
		URI = "https://api.safaricom.co.ke/"
	case "sandbox":
		URI = "https://sandbox.safaricom.co.ke/"
	default:
		return
	}
	client := &http.Client{}

	jsonValue, _ := json.Marshal(p)

	req, err := http.NewRequest(http.MethodPost, URI+"pulltransactions/v1/query", bytes.NewBuffer(jsonValue))

	if err != nil {
		return
	}
	req.Header.Add("Authorization", "Bearer "+token.Token)
	req.Header.Add("Content-Type", ContentType)
	resp, err := client.Do(req)

	if err != nil {
		return
	}
	defer resp.Body.Close()

	byteValue, _ := ioutil.ReadAll(resp.Body)

	return string(byteValue), err
}
