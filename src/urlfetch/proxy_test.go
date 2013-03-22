package urlfetch

import "testing"
import "fmt"

// Local helper functions
func test_getdata4urls(urls []string) {
    // create HTTP client
    client := HttpClient()

    ch := make(chan []byte)
    n := 0
    for _, url := range urls {
        n++
        go Getdata(client, url, ch)
    }
    for i:=0; i<n; i++ {
        fmt.Println(string(<-ch))
    }
}
func test_getdata(url string) {
    // create HTTP client
    client := HttpClient()

    ch := make(chan []byte)
    go Getdata(client, url, ch)
    fmt.Println(string(<-ch))
}

// Test function
func TestGetdata(t *testing.T) {
    url1 := "http://www.google.com"
    url2 := "http://www.golang.org"
    urls := []string{url1, url2}
    t.Log("test getdata call")
    test_getdata(url1)
    t.Log("test getdata call with multiple urls")
    test_getdata4urls(urls)
}
