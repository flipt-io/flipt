# shakedown

A tiny Bash DSL for HTTP testing with zero dependencies<sup>*</sup>.

Make HTTP requests and assert on the response body and headers.

<sub>* unless you count cURL and grep</sub>


## Example
Create `test.sh`:
```bash
#!/usr/bin/env bash
source shakedown.sh                             # load the framework

shakedown GET /about                            # make a GET request
  status 200                                    # assert response status is 200
  content_type 'text/html'                      # assert Content-Type header contains string
  header 'Expires'                              # assert response header exists containing string
  contains 'Take back your privacy!'            # assert response body contains string

shakedown POST / -d 'q=Shakedown'               # make a POST request with form data
  status 200
  contains 'Bob Seger'
```

Run the tests against a base URL:
```
$ ./test.sh -u https://duckduckgo.com
Starting shakedown of https://duckduckgo.com

GET /about
 ✔ status 200
 ✔ Content-Type: text/html
 ✔ header Expires
 ✔ contains "Take back your privacy!"

POST /
 ✔ status 200
 ✔ contains "Bob Seger"

Shakedown complete. 2 passed, 0 failed.
```


## DSL
```
shakedown <VERB> <PATH> <CURL OPTIONS>
  <assertion>
  <assertion>
  ...
```


## Assertions
```
status          <code>           Response status code = <code>
contains        <string>         Response body contains <string>
matches         <regex>          Response body matches <regex>
header          <string>         Response headers contains <string>
no_header       <string>         Response headers do not contain <string>
content_type    <string>         Content-Type header contains <string>
header_contains <name> <string>  Response header <name> contains <string>
```


## HTTP Authentication
Use the -c option to provide credentials.

```./test.sh -u my.domain.com -c user:pass```


## Setting cURL options
Any parameters after the path are passed straight on to cURL.

e.g. To send form data, follow redirects and set verbose output.

```shakedown POST /search -d 'query=shakedown' -L -v```


## Exit code
The exit code is set to the number of failed tests.


## Debugging
To help diagnose failing tests use ```print_headers```, ```print_body```, or make cURL verbose with '-v'.


## More Examples
```bash
#!/usr/bin/env bash
source shakedown.sh                               # load the framework

shakedown GET /foo                                # make a GET request
  status 404                                      # assert on http status code
  content_type 'text/html'                        # assert Content-Type header contains string
  contains 'Not found'                            # assert body contains string
  matches 'No.*'                                  # assert body matches regex

shakedown HEAD /                                  # make a HEAD request
  status 302

shakedown GET / -H 'Accept: application/json'     # add curl options
  print_headers                                   # output response headers for debugging
  print_body                                      # output response body for debugging
  status 200
  header 'Expires'

shakedown PUT /user/1 -d name=Rob                 # make a PUT request
  status 201

shakedown GET http://www.google.com -L            # provide full url to override default base url.
  status 200                                      # -L cURL option to follow redirects

shakedown GET http://www.google.com
  header_contains 'Referrer-Policy' 'no-referrer' # assert header 'Referrer-Policy' contains value 'no-referrer'
```


## Environment variables
The environment variables `SHAKEDOWN_URL` and `SHAKEDOWN_CREDENTIALS` can be used instead of passing -u and -c options.

```SHAKEDOWN_URL=https://duckduckgo.com ./test.sh```

Request timeouts can be set with:
```
SHAKEDOWN_CONNECT_TIMEOUT=5  # sets the curl option --connect-timeout. defaults to 5 seconds.
SHAKEDOWN_MAX_TIME=30        # sets the curl option --max-time. defaults to 30 seconds.
```

## Running tests in parallel
Divide your tests into multiple files, then run those in parallel, for example:

```bash
export SHAKEDOWN_URL=https://duckduckgo.com
ls -1 test-*.sh | parallel bash
```

## Docker

[![docker hub](https://img.shields.io/docker/cloud/build/robwhitby/shakedown.svg)](https://hub.docker.com/r/robwhitby/shakedown)

<https://hub.docker.com/r/robwhitby/shakedown>

```
docker run -v "$PWD":/work robwhitby/shakedown /work/sample-test.sh -u https://duckduckgo.com
```
