#!/usr/bin/env python
#-*- coding: utf-8 -*-
#pylint: disable-msg=
"""
File       : urls_client.py
Author     : Valentin Kuznetsov <vkuznet AT gmail dot com>
Description: 
"""

# system modules
import os
import sys
import urllib
import httplib
import urllib2

def main():
    "Main function"
    url1 = 'http://www.google.com'
    url2 = 'http://www.golang.org'
    params = {'urls': '\n'.join([url1, url2])}
    encoded_data = urllib.urlencode(params)
    url = "http://localhost:8000/getdata"
    req = urllib2.Request(url)
    handler = urllib2.HTTPHandler(debuglevel=1)
    opener  = urllib2.build_opener(handler)
    urllib2.install_opener(opener)
    data = urllib2.urlopen(req, encoded_data)
    info = data.info()
    code = data.getcode()
    print
    print code, info
    print data.read()

if __name__ == '__main__':
    main()
