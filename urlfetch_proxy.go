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
package main

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
func get_certs() []tls.Certificate {
    uproxy := os.Getenv("X509_USER_PROXY")
    uckey  := os.Getenv("X509_USER_KEY")
    ucert  := os.Getenv("X509_USER_CERT")
    tls_certs := []tls.Certificate{}
    if  len(uproxy) > 0 {
        x509cert, err := tls.LoadX509KeyPair(uproxy, uproxy)
        if  err != nil {
            fmt.Println("Fail to parser proxy X509 certificate", err)
            return []tls.Certificate{}
        }
        tls_certs = []tls.Certificate{x509cert}
    } else if len(uckey) > 0 {
        x509cert, err := tls.LoadX509KeyPair(ucert, uckey)
        if  err != nil {
            fmt.Println("Fail to parser user X509 certificate", err)
            return []tls.Certificate{}
        }
        tls_certs = []tls.Certificate{x509cert}
    } else {
        return []tls.Certificate{}
    }
    return tls_certs
}

/*
 * getdata(url string, ch chan<- []byte)
 * Fetches data for given URL and redirect response body to given channel
 */
func getdata(url string, ch chan<- []byte) {
    msg := ""
    certs := get_certs()
    if  len(certs) == 0 {
        msg = "Unable to fullfill your request, no server X509 certificate\n"
        ch <- []byte(msg)
        return
    }
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{tls.Certificates: certs},
    }
    client := &http.Client{Transport: tr}
    resp, err := client.Get(url)
    if  err != nil {
        msg = "Fail to contact " + url
        ch <- []byte(msg)
        return
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if  err != nil {
        msg := "Fail to parse reponse body"
        log.Println(msg, err)
        ch <- []byte(msg)
        return
    }
    ch <- body
}

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

/*
 * concurrent worker wrapper around getdata fucntion
 * it creates a channel and runs getdata concurrently to fetch data from given
 * url and stored them into the channel. All results are combined and copied
 * into output buffer.
 */
func getdata4urls(urls []string) []byte {
    ch := make(chan []byte)
    n := 0
    for _, url := range urls {
        n++
        go getdata(url, ch)
    }
    out := []byte{}
    for i:=0; i<n; i++ {
        out = AppendByte(out, <-ch)
    }
    return out
}

/*
 * RequestHandler is used by web server to handle incoming requests
 */
func RequestHandler(w http.ResponseWriter, r *http.Request) {
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

    // loop concurently over url list and store results into channel
    ch := make(chan []byte)
    n := 0
    for _, url := range urls {
        n++
        go getdata(url, ch)
    }
    // once channels are ready fill out results to response writer
    for i:=0; i<n; i++ {
        w.Write(<-ch)
        w.Write([]byte("\n"))
    }
}

func server(port string) {
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

/*
 * Test functions
 */
func test_getdata4urls(urls []string) {
    ch := make(chan []byte)
    n := 0
    for _, url := range urls {
        n++
        go getdata(url, ch)
    }
    for i:=0; i<n; i++ {
        fmt.Println(string(<-ch))
    }
}
func test_getdata(url string) {
    ch := make(chan []byte)
    go getdata(url, ch)
    fmt.Println(string(<-ch))
}
func test() {
    url1 := "http://www.google.com"
    url2 := "http://www.golang.org"
    urls := []string{url1, url2}
    fmt.Println("TEST: test_getdata")
    test_getdata(url1)
    fmt.Println("TEST: test_getdata4urls")
    test_getdata4urls(urls)
}

/*
 * MAIN
 */
func main() {
    server("8000")
}

