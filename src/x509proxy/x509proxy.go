package x509proxy
//package urlfetch

import "io/ioutil"
import "fmt"
import "regexp"
import "crypto/tls"

// Helper function to append bytes to existing slice
func AppendByte(slice []byte, data []byte) []byte {
    m := len(slice)
    n := m + len(data)
    if n > cap(slice) { // if necessary, reallocate
        // allocate double what's needed, for future growth.
        newSlice := make([]byte, (n+1)*2)
        copy(newSlice, slice)
        slice = newSlice
    }
    slice = slice[0:n]
    copy(slice[m:n], data)
    return slice
}

func getData(mkey string, block []byte) (keyBlock []byte) {
    newline := []byte("\n")
    out := []byte{}
    start := 0
    keyMatch := 0
    for i:=0; i<len(block); i++ {
        out = block[start:i]
        if  string(block[i]) == "\n" {
            test, _ := regexp.MatchString(mkey, string(out))
            if  test {
                keyMatch += 1
            }
            if  keyMatch > 0 {
                keyBlock = AppendByte(keyBlock, out)
                keyBlock = AppendByte(keyBlock, newline)
                if  keyMatch == 2 {
                    keyMatch = 0
                }
            }
            out = []byte{}
            start = i+1
        }
    }
    return
}
// LoadX509Proxy reads and parses a chained proxy file
// which contains PEN encoded data. It returns X509KeyPair.
func LoadX509Proxy(proxyFile string) (cert tls.Certificate, err error) {
        // read CERTIFICATE blocks
        certBlock, err := ioutil.ReadFile(proxyFile)
        if err != nil {
            return
        }
        certPEMBlock := getData("CERTIFICATE", certBlock)
        fmt.Println(string(certPEMBlock))

        // read KEY block
        keyBlock, err := ioutil.ReadFile(proxyFile)
        if err != nil {
            return
        }
        keyPEMBlock := getData("KEY", keyBlock)
        // test
        testBlock, err := ioutil.ReadFile("x509up_u502")
        fmt.Println(string(testBlock))
        if  string(testBlock) != string(keyPEMBlock) {
            fmt.Println("KEYS DIFFER", len(testBlock), len(keyPEMBlock))
            fmt.Println("testblock")
            fmt.Println(testBlock)
            fmt.Println("keyPEMBlock")
            fmt.Println(keyPEMBlock)
            newline := []byte("\n")
            fmt.Println("newline", newline)
        }

        return tls.X509KeyPair(certPEMBlock, keyPEMBlock)
}
