package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/skratchdot/open-golang/open"
	"html/template"
	"io/ioutil"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"flag"
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

func GetSettings() error {

	// See if we have a settings file
	dir, _ := os.Getwd()
	settingsFile := filepath.Join(dir, "settings.json")
	dat, err := ioutil.ReadFile(settingsFile)
	if err == nil {
		json.Unmarshal([]byte(dat), &s)
		log.Debug(s)
		return nil
	}

	// Client ID
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter clientId: ")
	s.ClientId, err = reader.ReadString('\n')
	if err != nil {
		return err
	}
	s.ClientId = strings.TrimSpace(s.ClientId)

	// Client Secret
	reader = bufio.NewReader(os.Stdin)
	fmt.Print("Enter clientSecret: ")
	s.ClientSecret, err = reader.ReadString('\n')
	if err != nil {
		return err
	}
	s.ClientSecret = strings.TrimSpace(s.ClientSecret)

	// Save settings to file
	jsonSettings, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(settingsFile, jsonSettings, 0644)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Loading template index.html")
	t, err := template.New("Index").Parse(IndexHTML)
	check(err)
	t.Execute(w, &s)
}

func getAuthCode(code string) (*accessToken, error) {
	uri := BaseMondoURL+"/oauth2/token"
	log.Debug(fmt.Sprintf("Fetching %s with code: %s", uri, code))
	resp, err := http.PostForm(uri,
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
	uri := BaseMondoURL+"/accounts"
	log.Debug(fmt.Sprintf("Fetching %s with token: %s", uri, authStruct.Access_token))
	req, err := http.NewRequest("GET", uri, nil)
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
	uri := BaseMondoURL+"/transactions"
	log.Debug(fmt.Sprintf("Fetching %s with token: %s", uri, authStruct.Access_token))
	req, err := http.NewRequest("GET", uri, nil)
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

		log.Debug(fmt.Sprintf("%+v\n", v))
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
	t, err := template.New("getTransactions").Parse(GetTransactionsHTML)
	check(err)
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

func SetLogLevel(level string){
	switch level {
	case "info":
	    log.SetLevel(log.InfoLevel)
	case "warn":
	    log.SetLevel(log.WarnLevel)
	case "debug":
	    log.SetLevel(log.DebugLevel)
	case "error":
	    log.SetLevel(log.ErrorLevel)
	case "fatal":
	    log.SetLevel(log.FatalLevel)
	case "panic":
			log.SetLevel(log.PanicLevel)
	default:
	    panic("unrecognized log level")
	}
}

func main() {
	level := flag.String("logLevel", "warn", "info, warn, debug, error, fatal, panic")
	flag.Parse()
	SetLogLevel(*level)

	log.Debug("Getting Settings")
	GetSettings()
	open.Run("http://localhost:8080/")
	http.HandleFunc("/", indexHandler)
	http.Handle("/files/", http.FileServer(http.Dir("")))
	http.HandleFunc("/getTransactions/", getTransactionsHandler)
	defer http.ListenAndServe(":8080", nil)
	log.Info("Running Webserver on localhost:8080")

}
