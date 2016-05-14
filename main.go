// Writing files in Go follows similar patterns to the
// ones we saw earlier for reading.

package main

import (
    "io/ioutil"
    "log"
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
  Transaction []Transaction `xml:"SIGNONMSGSRSV1>BANKMSGSRSV1>SMTRS>BANKTRANLIST>STMTTRN"`
}

func check(e error) {
    if e != nil {
        log.Fatal(e)
    }
}

func WriteXML(o *OFX, outputfile string) {


  output, err := xml.MarshalIndent(o, "  ", "    ")
  check(err)


  d1 := []byte(output)
  err = ioutil.WriteFile(outputfile, d1, 0644)
  check(err)
}

func main() {




}
