/*
 *
 * Author     : Valentin Kuznetsov <vkuznet AT gmail dot com>
 * Description: URL fetch proxy server concurrently fetches data from
 *              provided URL list. It provides a POST HTTP interface
 *              "/getdata" which accepts urls as newline separated encoded
 *              string
 * Created    : Wed Mar 20 13:29:48 EDT 2013
 * License    : MIT
 *
 */
package urlfetch

import (
    "os"
    "fmt"
    "log"
    "strings"
    "net/http"
    "io/ioutil"
    "crypto/tls"
)

/*
 * Return array of certificates
 */
func Certs() (tls_certs []tls.Certificate) {
    uproxy := os.Getenv("X509_USER_PROXY")
    uckey  := os.Getenv("X509_USER_KEY")
    ucert  := os.Getenv("X509_USER_CERT")
    if  len(uproxy) > 0 {
        x509cert, err := tls.LoadX509KeyPair(uproxy, uproxy)
        if  err != nil {
            fmt.Println("Fail to parser proxy X509 certificate", err)
            return
        }
        tls_certs = []tls.Certificate{x509cert}
    } else if len(uckey) > 0 {
        x509cert, err := tls.LoadX509KeyPair(ucert, uckey)
        if  err != nil {
            fmt.Println("Fail to parser user X509 certificate", err)
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
    if  len(certs) == 0 {
        client = &http.Client{}
        return
    }
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{tls.Certificates: certs},
    }
    client = &http.Client{Transport: tr}
    return
}

/*
 * Getdata(client *http.Client, url string, ch chan<- []byte)
 * Using given client and chanel fetches data for provided URL.
 * Results are redirected to specified channel
 */
func Getdata(client *http.Client, url string, ch chan<- []byte) {
    msg := ""
    resp, err := client.Get(url)
    if  err != nil {
        msg = "Fail to contact " + url
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
 * RequestHandler is used by web server to handle incoming requests
 */
func RequestHandler(w http.ResponseWriter, r *http.Request) {
    // we only accept POST request with urls (this is by design)
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
    log.Println(urls)

    // create HTTP client
    client := HttpClient()

    // loop concurently over url list and store results into channel
    ch := make(chan []byte)
    n := 0
    for _, url := range urls {
        n++
        go Getdata(client, url, ch)
    }
    // once channels are ready fill out results to response writer
    for i:=0; i<n; i++ {
        w.Write(<-ch)
        w.Write([]byte("\n"))
    }
}

// proxy server. It defines /getdata public interface
func Server(port string) {
    http.HandleFunc("/getdata", RequestHandler)
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

