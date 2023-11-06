# Static Token Authentication Example

This example shows how you can secure your Flipt instance with a static bootstrap token: <https://www.flipt.io/docs/configuration/authentication#method-static-token>

**Note:** You will not be able to use the Flipt UI with this example as it does not support static token authentication. 
See the [Dex example](../dex/README.md) to see how to configure Flipt to use OIDC for authentication and enable the UI.

## Requirements

To run this example application you'll need:

* [Docker](https://docs.docker.com/install/)
* [docker-compose](https://docs.docker.com/compose/install/)

## Running the Example

1. Run `docker-compose up` from this directory
1. Try to get a list of flags without authenticating using the REST API:

    ```shell
    ❯ curl -v http://localhost:8080/api/v1/flags

    > GET /api/v1/flags HTTP/1.1
    > Host: localhost:8080
    > Accept: */*
    >
    < HTTP/1.1 401 Unauthorized
    < Content-Type: text/plain; charset=utf-8
    ```

1. You should get a **401 Unauthorized** response as no authentication was present on the request
1. Try again, providing the bootstrap token `secret`, specified in the [docker-compose.yml](docker-compose.yml) file:

    ```shell
    ~ » curl -v  -H 'Authorization: Bearer secret' http://localhost:8080/api/v1/flags

    > GET /api/v1/flags HTTP/1.1
    > Host: localhost:8080
    > Accept: */*
    > Authorization: Bearer secret
    >
    < HTTP/1.1 200 OK
    < Content-Type: application/json
    < Grpc-Metadata-Content-Type: application/grpc
    < X-Content-Type-Options: nosniff
    < Content-Length: 46
    <
    {"flags":[],"nextPageToken":"","totalCount":0}
    ```

1. This time the request succeeds and a **200 OK** response is returned
