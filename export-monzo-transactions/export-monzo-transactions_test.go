package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
)

// Define tests
var xmltests = []struct {
	in  string
	out bool
}{
	{"<OFX>\\s*<BANKMSGSRSV1>\\s*<STMTRS>[\\S\\s]*<BANKTRANLIST>\\s*<STMTTRN>", true},
	{"</STMTTRN>\\s*</BANKTRANLIST>\\s*</STMTRS>\\s*</BANKMSGSRSV1>\\s*</OFX>", true},
	{"<NAME>HALFORDS 0371 MAIDENHEAD GB 0000</NAME>", true},
	{"<TRNAMT>-8.49</TRNAMT>", true},
	{"<NAME>GIANS RESTAURANT MAIDENHEAD GB 0000</NAME>", true},
	{"<TRNAMT>-14</TRNAMT>", true},
}

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	servUrl, _ := url.Parse(server.URL)
	BaseMonzoURL = servUrl.String()

	// Bake an auth request into the server to return a client
	mux.HandleFunc("/oauth2/token",
		func(w http.ResponseWriter, r *http.Request) {
			if val := r.FormValue("code"); val != "valid" {
				w.WriteHeader(401)
				return
			}
			fmt.Fprint(w, `{
											"access_token": "access_token",
											"client_id": "client_id",
											"expires_in": 21600,
											"refresh_token": "refresh_token",
											"token_type": "Bearer",
											"user_id": "user_id"
											}`)
		},
	)
}

func teardown() {
	mux = nil
	server = nil
}

func TestWriteXML(t *testing.T) {

	// Prepare a test OFX structure
	v := &OFX{}
	v.Transaction = append(v.Transaction, Transaction{
		TRNTYPE:  "POS",
		DTPOSTED: "20160408120000.000[+1]",
		TRNAMT:   -8.49,
		FITID:    "00POS201604081200000001-849HALFORDS0371MAIDENHEADGB0000",
		NAME:     "HALFORDS 0371 MAIDENHEAD GB 0000",
	})
	v.Transaction = append(v.Transaction, Transaction{
		TRNTYPE:  "PO123123S",
		DTPOSTED: "20160408120000.000[+1]",
		TRNAMT:   -14.0,
		FITID:    "00POS201604081200000001-140GIANSRESTAURANTMAIDENHEADGB0000",
		NAME:     "GIANS RESTAURANT MAIDENHEAD GB 0000",
	})

	// Get our CWD to write the ouput file to
	dir, err := os.Getwd()
	file := filepath.Join(dir, "test.ofx")

	// run WriteXML
	WriteXML(v, file)

	// Read the resulting file
	val, err := ioutil.ReadFile(file)
	check(err)

	// Check it contains what we expected
	for _, tt := range xmltests {
		result, _ := regexp.MatchString(tt.in, string(val))
		if result != tt.out {
			t.Fatalf("expected output of '%s' got '%s'", tt.in, string(val))
		}
	}
}

func TestOAuth(t *testing.T) {
	setup()
	defer teardown()

	client, err := getAuthCode("valid")
	assert.NoError(t, err)

	assert.NotNil(t, client)
	assert.NotNil(t, client.Access_token)
	assert.NotNil(t, client.Expires_in)
	assert.NotNil(t, client.Refresh_token)
	assert.NotNil(t, client.Token_type)

	client, err = getAuthCode("invalid")
	assert.Error(t, err)

}

func TestTransactions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{
  "accounts": [
    {
      "id": "acc_000097rJJuKs0XcJLnVzTW",
      "created": "2016-05-04T13:50:41.289Z",
      "description": "Sam Martin"
    }
  ]
}
`)
		},
	)

	mux.HandleFunc("/transactions",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{"transactions": [{
			            "account_balance": 13013,
			            "amount": -510,
			            "created": "2015-08-22T12:20:18Z",
			            "currency": "GBP",
			            "description": "THE DE BEAUVOIR DELI C LONDON        GBR",
			            "id": "tx_00008zIcpb1TB4yeIFXMzx",
									"merchant": {
											"address": {
												"address": "98 Southgate Road",
												"city": "London",
												"country": "GB",
												"latitude": 51.54151,
												"longitude": -0.08482400000002599,
												"postcode": "N1 3JD",
												"region": "Greater London"
											},
											"created": "2015-08-22T12:20:18Z",
											"group_id": "grp_00008zIcpbBOaAr7TTP3sv",
											"id": "merch_00008zIcpbAKe8shBxXUtl",
											"logo": "https://pbs.twimg.com/profile_images/527043602623389696/68_SgUWJ.jpeg",
											"emoji": "🍞",
											"name": "The De Beauvoir Deli Co.",
											"category": "eating_out"
										},
			            "metadata": {},
			            "notes": "Salmon sandwich 🍞",
			            "is_load": false,
			            "settled": true,
			            "category": "eating_out"
			        },
			        {
			            "account_balance": 12334,
			            "amount": -679,
			            "created": "2015-08-23T16:15:03Z",
			            "currency": "GBP",
			            "description": "VUE BSL LTD            ISLINGTON     GBR",
			            "id": "tx_00008zL2INM3xZ41THuRF3",
									"merchant": {
											"address": {
												"address": "98 Southgate Road",
												"city": "London",
												"country": "GB",
												"latitude": 51.54151,
												"longitude": -0.08482400000002599,
												"postcode": "N1 3JD",
												"region": "Greater London"
											},
											"created": "2015-08-22T12:20:18Z",
											"group_id": "grp_00008zIcpbBOaAr7TTP3sv",
											"id": "merch_00008zIcpbAKe8shBxXUtl",
											"logo": "https://pbs.twimg.com/profile_images/527043602623389696/68_SgUWJ.jpeg",
											"emoji": "🍞",
											"name": "The De Beauvoir Deli Co.",
											"category": "eating_out"
										},
			            "metadata": {},
			            "notes": "",
			            "is_load": false,
			            "settled": true,
			            "category": "eating_out"
			        }]}
`)
		},
	)

	client, err := getAuthCode("valid")
	assert.NoError(t, err)

	accounts, err := getAccounts(client.Access_token, client.User_id )
	assert.NoError(t, err)
	account := accounts.Accounts[0]
	assert.Exactly(t, account.Id, "acc_000097rJJuKs0XcJLnVzTW")

	transactions, err := getTransactions(account.Id, client.Access_token)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(transactions.Transactions))
	assert.Equal(t, transactions.Transactions[0].Currency, "GBP")
	assert.Equal(t, transactions.Transactions[0].Amount, -510)
}
