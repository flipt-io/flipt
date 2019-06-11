#!/usr/bin/env bash

# To run:
# ./sample-test.sh -u https://duckduckgo.com
# or
# SHAKEDOWN_URL=https://duckduckgo.com ./sample-test.sh

source shakedown.sh                         # load the framework

shakedown GET /about                        # make a GET request
  status 200                                # assert response status is 200
  content_type 'text/html'                  # assert content type contains string
  header 'Expires'                          # assert response header exists containing string
  contains 'You deserve privacy'            # assert resposne body contains string

shakedown POST / -d 'q=Shakedown'           # make a POST request with form data
  status 200
  contains 'Bob Seger'

shakedown GET http://www.google.com -L      # provide full url to override default base url.
  status 200                                # -L cURL option to follow redirects

shakedown GET http://www.google.com
  header_contains 'Cache-Control' 'private' # assert header 'Cache-Control' contains 'private'
  #print_headers                            # debug helper
  #print_body                               # debug helper
