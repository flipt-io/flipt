<p align="center">
    <img src="../../images/dex.png" alt="Dex Login" />
</p>

# OIDC Authentication with Dex

This is a demonstration of using Flipt + Dex as an OIDC provider for authentication.

## Requirements

To run this example application you'll need:

* [Docker](https://docs.docker.com/install/)
* [docker-compose](https://docs.docker.com/compose/install/)

## Running

1. You're going to need an entry for `dex` in your `/etc/hosts` for this configuration.

    This is so that we can support setups such as "Docker for Mac", which don't have full "host" network mode for Docker.

    Add the following to your `/etc/hosts` file on your host machine:

    ```text
    127.0.0.1 dex
    ```

    e.g. `sudo sh -c 'echo "127.0.0.1 dex" >> /etc/hosts'`.

1. Run `docker-compose up` from this directory.
1. Navigate to [Flipt in your browser](http://localhost:8080).
1. Click 'Login With Dex'
1. Login using email: `admin@example.com` and password: `password`.
1. Select `Grant Access`.

From here you should be navigated back to Flipt and an authenticated session should be established.
