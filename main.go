// Writing files in Go follows similar patterns to the
// ones we saw earlier for reading.

package main

import (
  "io/ioutil"
  "bufio"
  "os"
  "log"
  "fmt"
  "encoding/xml"
  "net/http"
	//"github.com/gorilla/mux"
  "encoding/json"
  "html/template"
  "strings"
  "net/url"
  "path/filepath"
  "time"
)

type Transaction struct {
  TRNTYPE   string
  DTPOSTED  string
  TRNAMT    float32
  FITID     string
  NAME      string
}

type BankAccount struct {
  BANKID   string
  ACCTID   string
  ACCTTYPE string
}

type OFX struct {
  XMLName     xml.Name      `xml: "OFX"`
  BankAccount BankAccount   `xml:"BANKMSGSRSV1>STMTRS>BANKACCTFROM"`
  Transaction []Transaction `xml:"BANKMSGSRSV1>STMTRS>BANKTRANLIST>STMTTRN"`
}

type settings struct {
  ClientId string
  ClientSecret string
}

type accessToken struct {
  Access_token string
  Client_id string
  Expires_in string
  Refresh_token string
  Token_type string
  User_id string
}

type account struct {
  Id string
  Created string
  Description string
}

type accounts struct {
  Accounts []account
}

type transaction struct {
	Account_balance int
	Amount         int
	Attachments    []interface{}
	Category       string
	Created        string
	Currency       string
	Description    string
	ID             string
	Is_load         bool
	Merchant       string
	Metadata       map[string]interface{}
	Notes          string
	Settled        string
}

type transactions struct {
  Transactions []transaction
}

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

func getAuthCode(code string) accessToken  {
  resp, err := http.PostForm("https://api.getmondo.co.uk/oauth2/token",
	   url.Values{
        "grant_type": {"authorization_code"},
        "client_id": {s.ClientId},
        "client_secret": {s.ClientSecret},
        "redirect_uri": {"http://localhost:8080/getTransactions/"},
        "code": {code},
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
  req.Header.Add("authorization", `Bearer ` + authStruct.Access_token)
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

func getTransactions(authStruct accessToken, acccountStruct account) transactions{
  // Fetch transactions
  client := &http.Client{}
  req, err := http.NewRequest("GET", "https://api.getmondo.co.uk/transactions", nil)
  check(err)
  req.Header.Add("authorization", `Bearer ` + authStruct.Access_token)
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

func getTransactionsHandler(w http.ResponseWriter, r *http.Request){

  // Authenticate
  code := r.FormValue("code")
  authStruct := getAuthCode(code)

  accounts := getAccounts(authStruct)
  account := accounts.Accounts[0]

  transactions := getTransactions(authStruct, account)

  // Create an OFX
  OFXStruct := &OFX{}
  OFXStruct.BankAccount = BankAccount{
    BANKID: "0",
    ACCTID: account.Id,
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
    var amount = float32(v.Amount)/float32(100)
    OFXStruct.Transaction = append(OFXStruct.Transaction,Transaction{
      TRNTYPE: "POS",
      DTPOSTED: formattedTime,
      TRNAMT: amount,
      FITID: v.ID,
      NAME: v.Description,
    })
  }

  // Get our CWD to write the ouput file to
  dir, err := os.Getwd()
  check(err)
  file := filepath.Join(dir, "output.ofx")

  // run WriteXML
  WriteXML(OFXStruct, file);


  t, _ := template.ParseFiles("getTransactions.html")
  t.Execute(w, &s)
}

func WriteXML(o *OFX, outputfile string) {

  output, err := xml.MarshalIndent(o, "  ", "    ")
  check(err)
  d1 := []byte(output)
  err = ioutil.WriteFile(outputfile, d1, 0644)
  check(err)
}

func main() {
  // getsettings outputs to global variable
  getsettings()
  //rtr := mux.NewRouter()
  http.HandleFunc("/", indexHandler)
  http.HandleFunc("/getTransactions/", getTransactionsHandler)
  http.ListenAndServe(":8080", nil)
  log.Print("Started webserver on 8080");


}
