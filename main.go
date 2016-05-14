// Writing files in Go follows similar patterns to the
// ones we saw earlier for reading.

package main

import (
    "io/ioutil"
    "log"
    "path/filepath"
    "os"
    "encoding/xml"
)

type Transaction struct{
  //XMLName   xml.Name `xml:"STMTTRN"`
  TRNTYPE   string
  DTPOSTED  string
  TRNAMT    float32
  FITID     string
  NAME      string
}

type OFX struct {
  XMLName     xml.Name `xml: "OFX"`
  Transaction Transaction `xml:"SIGNONMSGSRSV1>BANKMSGSRSV1>SMTRS>BANKTRANLIST>STMTTRN"`
}

func check(e error) {
    if e != nil {
        log.Fatal(e)
    }
}

func main() {

    // Get our CWD
    dir, err := os.Getwd()

    v := &OFX{}
    v.Transaction = Transaction{
      TRNTYPE: "POS",
      DTPOSTED: "20160408120000.000[+1]",
      TRNAMT: -8.49,
      FITID: "00POS201604081200000001-849HALFORDS0371MAIDENHEADGB0000",
      NAME: "HALFORDS 0371 MAIDENHEAD GB 0000",
    }

    output, err := xml.MarshalIndent(v, "  ", "    ")
    check(err)

    // To start, here's how to dump a string (or just
    // bytes) into a file.
    d1 := []byte(output)
    err = ioutil.WriteFile(filepath.Join(dir,"temp.txt"), d1, 0644)
    check(err)


}
