package main

import (
  "testing"
  "os"
  "path/filepath"
  "io/ioutil"
  "regexp"
)

// Define tests
var xmltests = []struct {
  in  string
  out bool
}{
  {"<OFX>\\s*<SIGNONMSGSRSV1>\\s*<BANKMSGSRSV1>\\s*<SMTRS>\\s*<BANKTRANLIST>\\s*<STMTTRN>", true},
  {"</STMTTRN>\\s*</BANKTRANLIST>\\s*</SMTRS>\\s*</BANKMSGSRSV1>\\s*</SIGNONMSGSRSV1>\\s*</OFX>",true},
  {"<NAME>HALFORDS 0371 MAIDENHEAD GB 0000</NAME>", true},
  {"<TRNAMT>-8.49</TRNAMT>", true},
  {"<NAME>GIANS RESTAURANT MAIDENHEAD GB 0000</NAME>", true},
  {"<TRNAMT>-14</TRNAMT>", true},
}

func TestWriteXML(t *testing.T){

  // Prepare a test OFX structure
  v := &OFX{}
  v.Transaction = append(v.Transaction,Transaction{
    TRNTYPE: "POS",
    DTPOSTED: "20160408120000.000[+1]",
    TRNAMT: -8.49,
    FITID: "00POS201604081200000001-849HALFORDS0371MAIDENHEADGB0000",
    NAME: "HALFORDS 0371 MAIDENHEAD GB 0000",
  })
  v.Transaction = append(v.Transaction,Transaction{
    TRNTYPE: "PO123123S",
    DTPOSTED: "20160408120000.000[+1]",
    TRNAMT: -14.0,
    FITID: "00POS201604081200000001-140GIANSRESTAURANTMAIDENHEADGB0000",
    NAME: "GIANS RESTAURANT MAIDENHEAD GB 0000",
  })

  // Get our CWD to write the ouput file to
  dir, err := os.Getwd()
  file := filepath.Join(dir, "test.ofx")

  // run WriteXML
  WriteXML(v, file);

  // Read the resulting file
  val, err := ioutil.ReadFile(file)
  check(err)

  // Check it contains what we expected
  for _, tt := range xmltests {
    result, _ := regexp.MatchString(tt.in, string(val))
    if result != tt.out {
      t.Fatalf("expected output of '%s' to be %s", tt.in, string(val))
    }
  }
}
