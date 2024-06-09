# Reverse Proxy: OIDC Authentication

This example shows how you can secure your Flipt instance using OIDC Authentication behind a reverse proxy using [nginx](https://www.nginx.com/).

This example uses [Dex](https://dexidp.io/) as the OIDC provider and takes advantage of the `X-Forwarded-Prefix` header to ensure that the OIDC callback URL is correctly set.

## Requirements

To run this example application you'll need:

* [Docker](https://docs.docker.com/install/)
* [docker-compose](https://docs.docker.com/compose/install/)

## Running the Example

1. Run `docker-compose up` from this directory
1. Open the example UI (default: [http://localhost:8080](http://localhost:8080))
1. Click `Go to Flipt` to be redirected to the Flipt UI
1. Click `Login with Dex` to be redirected to the Dex login page
1. Use `admin@example.com` and `password` for authentication in Dex 
1. Notice that Flipt is being served at `/flipt` instead of the default root path
