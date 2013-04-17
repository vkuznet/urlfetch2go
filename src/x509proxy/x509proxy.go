package x509proxy

import "io/ioutil"
import "fmt"
import "regexp"
import "crypto/tls"
import "crypto/x509"
import "encoding/pem"
import "errors"
import "crypto/rsa"

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
// which contains PEM encoded data. It returns X509KeyPair.
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

//        return tls.X509KeyPair(certPEMBlock, keyPEMBlock)
        return X509KeyPair(certPEMBlock, keyPEMBlock)
}

// X509KeyPair parses a public/private key pair from a pair of
// PEM encoded data.
func X509KeyPair(certPEMBlock, keyPEMBlock []byte) (cert tls.Certificate, err error) {
    fmt.Println("CALL local X509KeyPair")
	var certDERBlock *pem.Block
	for {
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
        // parse certificates
        certs, err2 := x509.ParseCertificates(certDERBlock.Bytes)
        cert.Leaf = certs[0]
        fmt.Println("ParseCertificates", certs, len(certs), err2)
        if certDERBlock.Type == "CERTIFICATE" {
            cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
        }
	}
    fmt.Println("cert.Certificate, len=", len(cert.Certificate))

	if len(cert.Certificate) == 0 {
		err = errors.New("crypto/tls: failed to parse certificate PEM data")
		return
	}

	keyDERBlock, _ := pem.Decode(keyPEMBlock)
	if keyDERBlock == nil {
		err = errors.New("crypto/tls: failed to parse key PEM data")
		return
	}

	// OpenSSL 0.9.8 generates PKCS#1 private keys by default, while
	// OpenSSL 1.0.0 generates PKCS#8 keys. We try both.
	var key *rsa.PrivateKey
	if key, err = x509.ParsePKCS1PrivateKey(keyDERBlock.Bytes); err != nil {
		var privKey interface{}
		if privKey, err = x509.ParsePKCS8PrivateKey(keyDERBlock.Bytes); err != nil {
			err = errors.New("crypto/tls: failed to parse key: " + err.Error())
			return
		}

		var ok bool
		if key, ok = privKey.(*rsa.PrivateKey); !ok {
			err = errors.New("crypto/tls: found non-RSA private key in PKCS#8 wrapping")
			return
		}
	}

	cert.PrivateKey = key
    fmt.Println("CALL local X509KeyPair ended")

	return
}
