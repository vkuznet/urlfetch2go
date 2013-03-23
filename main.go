package main

import "urlfetch"
import "flag"

func main() {
    var port string
    flag.StringVar(&port, "port", "8000", "URL fetch server port number")
    flag.Parse()
    urlfetch.Server(port)
}
