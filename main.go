package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/skratchdot/open-golang/open"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	BaseMondoURL              = "https://api.getmondo.co.uk"
	s                         settings
	ErrUnauthenticatedRequest = fmt.Errorf("your request was not sent with a valid token")
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func getsettings() {
	// Client ID
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter clientId: ")
	s.ClientId, _ = reader.ReadString('\n')
	s.ClientId = strings.TrimSpace(s.ClientId)

	// Client Secret
	reader = bufio.NewReader(os.Stdin)
	fmt.Print("Enter clientSecret: ")
	s.ClientSecret, _ = reader.ReadString('\n')
	s.ClientSecret = strings.TrimSpace(s.ClientSecret)

}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.html")
	t.Execute(w, &s)
}

func getAuthCode(code string) (*accessToken, error) {
	resp, err := http.PostForm(BaseMondoURL+"/oauth2/token",
		url.Values{
			"grant_type":    {"authorization_code"},
			"client_id":     {s.ClientId},
			"client_secret": {s.ClientSecret},
			"redirect_uri":  {"http://localhost:8080/getTransactions/"},
			"code":          {code},
		},
	)
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		return nil, ErrUnauthenticatedRequest
	}
	body, err := ioutil.ReadAll(resp.Body)
	var result accessToken
	json.Unmarshal([]byte(body), &result)
	return &result, err
}

func getAccounts(authStruct *accessToken) (*accounts, error) {
	// Prepare HTTP request
	client := &http.Client{}
	req, err := http.NewRequest("GET", BaseMondoURL+"/accounts", nil)
	check(err)
	req.Header.Add("authorization", `Bearer `+authStruct.Access_token)
	q := req.URL.Query()
	q.Add("account_id", authStruct.User_id)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		return nil, ErrUnauthenticatedRequest
	}
	body, err := ioutil.ReadAll(resp.Body)

	// JSON decode result
	var result accounts
	json.Unmarshal([]byte(body), &result)

	return &result, nil
}

func getTransactions(authStruct *accessToken, acccountStruct account) (*transactions, error) {
	// Fetch transactions
	client := &http.Client{}
	req, err := http.NewRequest("GET", BaseMondoURL+"/transactions", nil)
	check(err)
	req.Header.Add("authorization", `Bearer `+authStruct.Access_token)
	q := req.URL.Query()
	q.Add("account_id", acccountStruct.Id)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		return nil, ErrUnauthenticatedRequest
	}
	body, err := ioutil.ReadAll(resp.Body)

	// JSON decode result
	var result transactions
	json.Unmarshal([]byte(body), &result)

	return &result, nil
}

func getTransactionsHandler(w http.ResponseWriter, r *http.Request) {

	// Authenticate
	code := r.FormValue("code")
	authStruct, err := getAuthCode(code)
	check(err)

	// Find our accout ID
	accounts, err := getAccounts(authStruct)
	check(err)
	account := accounts.Accounts[0]

	// Fetch transaction
	transactions, err := getTransactions(authStruct, account)
	check(err)

	// Create an OFX
	OFXStruct := &OFX{}
	OFXStruct.BankAccount = BankAccount{
		BANKID:   "0",
		ACCTID:   account.Id,
		ACCTTYPE: "CHECKING",
	}

	// Loop through the transactions adding them to the OFX struct
	for _, v := range transactions.Transactions {

		// Exclude 0 value transactions (e.g. pin resets)
		if v.Amount == 0 {
			continue
		}

		// Format the time
		time, _ := time.Parse(time.RFC3339, v.Created)
		formattedTime := time.Format("20060102150405.000[-07]")

		// Convert the transaction to a float in Pounds instead of an INT in pennies
		var amount = float32(v.Amount) / float32(100)

		// Save the transaction
		OFXStruct.Transaction = append(OFXStruct.Transaction, Transaction{
			TRNTYPE:  "POS",
			DTPOSTED: formattedTime,
			TRNAMT:   amount,
			FITID:    v.ID,
			NAME:     v.Description,
		})
	}

	// Get our CWD to write the ouput file to
	dir, err := os.Getwd()
	check(err)

	// Get the current time to create a unique filename
	timeNow := time.Now()
	fileName := timeNow.Format("2006-01-02T15-04-05.999999") + ".ofx"
	os.MkdirAll(filepath.Join(dir, "files"), 0644)
	fileAbsolute := filepath.Join(dir, "files", fileName)

	// Save to file
	WriteXML(OFXStruct, fileAbsolute)

	// Show the web page
	t, _ := template.ParseFiles("getTransactions.html")
	getTransactionsStruct := &getTransactionsTemplateVars{
		FileAbsolute: fileAbsolute,
		FileName:     fileName,
	}

	t.Execute(w, &getTransactionsStruct)
}

func WriteXML(o *OFX, outputfile string) {

	output, err := xml.MarshalIndent(o, "  ", "    ")
	check(err)
	d1 := []byte(output)
	err = ioutil.WriteFile(outputfile, d1, 0644)
	check(err)
}

func main() {
	getsettings()
	open.Run("http://localhost:8080/")
	http.HandleFunc("/", indexHandler)
	http.Handle("/files/", http.FileServer(http.Dir("")))
	http.HandleFunc("/getTransactions/", getTransactionsHandler)
	defer http.ListenAndServe(":8080", nil)
	log.Print("Running Webserver on localhost:8080")

}
