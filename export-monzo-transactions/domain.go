package main

import (
	"encoding/xml"
)

type Transaction struct {
	TRNTYPE  string
	DTPOSTED string
	TRNAMT   float32
	FITID    string
	NAME     string
	RunningTotal string `xml:",comment"`
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
	ClientId     string
	ClientSecret string
}

type accessToken struct {
	Access_token  string
	Client_id     string
	Expires_in    string
	Refresh_token string
	Token_type    string
	User_id       string
}

type account struct {
	Id          string
	Created     string
	Description string
}

type accounts struct {
	Accounts []account
}

type transaction struct {
	Account_balance int
	Amount          int
	Attachments     []interface{}
	Category        string
	Created         string
	Currency        string
	Description     string
	ID              string
	Is_load         bool
	Merchant        string
	Metadata        map[string]interface{}
	Notes           string
	Settled         string
	Local_Amount	  int
	Local_Currency  string
	Updated         string
	Account_ID      string
	Counterparty    map[string]interface{}
	Scheme          string
	Dedupe_ID       string
	Originator      bool
	Decline_Reason  string

}

type transactions struct {
	Transactions []transaction
}

type getTransactionsTemplateVars struct {
	Accounts     []account
	AccessToken  string
	UserId       string
}
