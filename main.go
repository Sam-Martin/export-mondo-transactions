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

var s settings

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

func getAuthCode(code string) accessToken {
	resp, err := http.PostForm("https://api.getmondo.co.uk/oauth2/token",
		url.Values{
			"grant_type":    {"authorization_code"},
			"client_id":     {s.ClientId},
			"client_secret": {s.ClientSecret},
			"redirect_uri":  {"http://localhost:8080/getTransactions/"},
			"code":          {code},
		},
	)
	check(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	check(err)
	var result accessToken
	json.Unmarshal([]byte(body), &result)
	return result
}

func getAccounts(authStruct accessToken) accounts {
	// Prepare HTTP request
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.getmondo.co.uk/accounts", nil)
	check(err)
	req.Header.Add("authorization", `Bearer `+authStruct.Access_token)
	q := req.URL.Query()
	q.Add("account_id", authStruct.User_id)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// JSON decode result
	var result accounts
	json.Unmarshal([]byte(body), &result)

	return result
}

func getTransactions(authStruct accessToken, acccountStruct account) transactions {
	// Fetch transactions
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.getmondo.co.uk/transactions", nil)
	check(err)
	req.Header.Add("authorization", `Bearer `+authStruct.Access_token)
	q := req.URL.Query()
	q.Add("account_id", acccountStruct.Id)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// JSON decode result
	var result transactions
	json.Unmarshal([]byte(body), &result)

	return result
}

func getTransactionsHandler(w http.ResponseWriter, r *http.Request) {

	// Authenticate
	code := r.FormValue("code")
	authStruct := getAuthCode(code)

	accounts := getAccounts(authStruct)
	account := accounts.Accounts[0]

	transactions := getTransactions(authStruct, account)

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
		log.Print(string(v.Description))

		time, _ := time.Parse(time.RFC3339, v.Created)
		formattedTime := time.Format("20060102150405.000[-07]")
		var amount = float32(v.Amount) / float32(100)
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
	fileAbsolute := filepath.Join(dir, "files", fileName)

	// run WriteXML
	WriteXML(OFXStruct, fileAbsolute)

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
	http.ListenAndServe(":8080", nil)

}
