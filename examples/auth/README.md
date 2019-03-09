# Authentication Example

This example shows how you can secure your Flipt instance using HTTP Basic Authentication.

It is important to note that Flipt currently has no built in authentication, authorization or encryption as Flipt was designed to work inside your trusted architecture and not be exposed publicly.

That being said, if you do want to expose the Flipt dashboard and REST API publicly using HTTP Basic Authentication, you can do so by using a reverse proxy. This example uses the [Caddy](https://caddyserver.com/) web server to accomplish this task.

## Limitations

Currently there is no way provided in Flipt to configure authentication with the GRPC API if exposed publicly. Thus, if you choose this approach is is important to **not expose the GRPC port publicly!** as anyone would be able to access your resources over GRPC without authenticating.

## Requirements

To run this example application you'll need:

* [Docker](https://docs.docker.com/install/)
* [docker-compose](https://docs.docker.com/compose/install/)

## Running the Example

1. Run `docker-compose up` from this directory
1. Open the Flipt UI (default: [http://localhost:8080](http://localhost:8080))
1. Note you will be prompted for a username and password, the default is `admin:password`. This can be configured by setting the `HTTP_USERNAME` and `HTTP_PASSWORD` environment variables in the [docker-compose.yml](docker-compose.yml) file.
1. Try to get a list of flags without authenticating using the REST API:
    ```shell
    ❯ curl -v http://localhost:8080/api/v1/flags
    *   Trying 127.0.0.1...
    * TCP_NODELAY set
    * Connected to localhost (127.0.0.1) port 8080 (#0)
    > GET /api/v1/flags HTTP/1.1
    > Host: localhost:8080
    > User-Agent: Mozilla/5.0 (Windows NT 6.1; rv:45.0) Gecko/20100101 Firefox/45.0
    > Accept: */*
    > Referer:
    >
    < HTTP/1.1 401 Unauthorized
    < Content-Type: text/plain; charset=utf-8
    ```
1. You should get a **401 Unauthorized** response as no username or password was present on the request
1. Try again, providing the username and password:
    ```shell
    ❯ curl -v -u admin:password http://localhost:8080/api/v1/flags
    *   Trying 127.0.0.1...
    * TCP_NODELAY set
    * Connected to localhost (127.0.0.1) port 8080 (#0)
    * Server auth using Basic with user 'admin'
    > GET /api/v1/flags HTTP/1.1
    > Host: localhost:8080
    > Authorization: Basic YWRtaW46cGFzc3dvcmQ=
    > User-Agent: Mozilla/5.0 (Windows NT 6.1; rv:45.0) Gecko/20100101 Firefox/45.0
    > Accept: */*
    > Referer:
    >
    < HTTP/1.1 200 OK
    < Content-Type: application/json
    ```
1. This time the request succeeds and a **200 OK** response is returned
