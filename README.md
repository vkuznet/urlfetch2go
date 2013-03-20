urlfetch2go
===========

URL fetch proxy server (Go implementation)

### Description

This is draft implementation of URL fetch proxy server. The main idea is to
have a proxy server for concurent data fetching from provided URL lists. It is
implemented as sinle HTTP server which accepts POST request from a client. The
POST method is choosen due to ability to send multiple (large) URLs to proxy
server for data retrieval.

### Installation & Usage

To compile the URL server you need a Go compiler, then perform the following:

```
go build urlfetch_proxy.go
```

It will build ```urlfetch_proxy``` executable which you can fetch from UNIX shell.
By default it serves request on port 8000, feel free to modify code accoringly.

I also provided a simple python client to demonstrate the usage of proxy
server. Modify code accoringly for your favorite URLs and run it as following:

```
python ./urlfetch_client.py
```

