/*
 *
 * Author     : Valentin Kuznetsov <vkuznet AT gmail dot com>
 * Description: URL fetch proxy server concurrently fetches data from
 *              provided URL list. It provides a POST HTTP interface
 *              "/fetch" which accepts urls as newline separated encoded
 *              string
 * Created    : Wed Mar 20 13:29:48 EDT 2013
 * License    : MIT
 *
 */
package urlfetch

import (
    "os"
    "log"
    "strings"
    "net/http"
    "io/ioutil"
    "crypto/tls"
    "regexp"
    "x509proxy"
)

/*
 * Return array of certificates
 */
func Certs() (tls_certs []tls.Certificate) {
    uproxy := os.Getenv("X509_USER_PROXY")
    uckey  := os.Getenv("X509_USER_KEY")
    ucert  := os.Getenv("X509_USER_CERT")
    log.Println("X509_USER_PROXY", uproxy)
    log.Println("X509_USER_KEY", uckey)
    log.Println("X509_USER_CERT", ucert)
    if  len(uproxy) > 0 {
        // use local implementation of LoadX409KeyPair instead of tls one
        x509cert, err := x509proxy.LoadX509Proxy(uproxy)
        if  err != nil {
            log.Println("Fail to parser proxy X509 certificate", err)
            return
        }
        tls_certs = []tls.Certificate{x509cert}
    } else if len(uckey) > 0 {
        x509cert, err := tls.LoadX509KeyPair(ucert, uckey)
        if  err != nil {
            log.Println("Fail to parser user X509 certificate", err)
            return
        }
        tls_certs = []tls.Certificate{x509cert}
    } else {
        return
    }
    return
}

/*
 * HTTP client for urlfetch server
 */
func HttpClient() (client *http.Client) {
    // create HTTP client
    certs := Certs()
    log.Println("Number of certificates", len(certs))
    if  len(certs) == 0 {
        client = &http.Client{}
        return
    }
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{tls.Certificates: certs,
                                     tls.InsecureSkipVerify: true},
    }
    log.Println("Create TLSClientConfig")
    client = &http.Client{Transport: tr}
    return
}

// create global HTTP client and re-use it through the code
var client = HttpClient()

/*
 * Fetch(url string, ch chan<- []byte)
 * Fetch data for provided URL and redirect results to given channel
 */
func Fetch(url string, ch chan<- []byte) {
    msg := ""
    if  validate_url(url) == false {
        ch <- []byte(msg)
        return
    }
    resp, err := client.Get(url)
    if  err != nil {
        msg = "Fail to contact " + url
        log.Println(msg, err)
        ch <- []byte(msg)
        return
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if  err != nil {
        msg = "Fail to parse reponse body"
        log.Println(msg, err)
        ch <- []byte(msg)
        return
    }
    ch <- body
}

/*
 * Helper function which validates given URL
 */
func validate_url(url string) (bool) {
    if  len(url) > 0 {
        pat := "(https|http)://[-A-Za-z0-9_+&@#/%?=~_|!:,.;]*[-A-Za-z0-9+&@#/%=~_|]"
        matched, err := regexp.MatchString(pat, url)
        if err == nil {
            if  matched == true {
                return true
            }
        }
        log.Println("ERROR invalid URL:", url)
    }
    return false
}

/*
 * RequestHandler is used by web server to handle incoming requests
 */
func RequestHandler(w http.ResponseWriter, r *http.Request) {
    // we only accept POST requests with urls (this is by design)
    if  r.Method != "POST" {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    // parse input request parameter, in this case we should pass urls
    r.ParseForm()
    urls := []string{}
    for k, v := range r.Form {
        if k == "urls" {
            urls = strings.Split(v[0], "\n")
        }
    }
    num := len(urls)
    if  num > 0 {
        first := urls[0]
        last := urls[num-1]
        log.Println("Fetch", num, "URLs:", first, "...", last)
    } else {
        w.Write([]byte("No URLs provided\n"))
        return
    }

    // loop concurently over url list and store results into channel
    ch := make(chan []byte)
    for _, url := range urls {
        go Fetch(url, ch)
    }
    // once channels are ready fill out results to response writer
    for i:=0; i<len(urls); i++ {
        w.Write(<-ch)
        w.Write([]byte("\n"))
    }
}

// proxy server. It defines /fetch public interface
func Server(port string) {
    log.Printf("Start server localhost:%s/fetch", port)
    http.HandleFunc("/fetch", RequestHandler)
    err := http.ListenAndServe(":" + port, nil)
    // NOTE: later this can be replaced with secure connection
    // replace ListenAndServe(addr string, handler Handler)
    // with TLS function
    // ListenAndServeTLS(addr string, certFile string, keyFile string, handler
    // Handler)
    if  err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

