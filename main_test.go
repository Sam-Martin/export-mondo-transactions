package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"net/http"
	"net/http/httptest"
	"net/url"
	"fmt"
	"github.com/stretchr/testify/assert"
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
	BaseMondoURL = servUrl.String()

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
			t.Fatalf("expected output of '%s' to be '%s' got '%s'", tt.in, tt.out, string(val))
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
