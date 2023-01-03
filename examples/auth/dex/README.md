Flipt + Dex
-----------

This is a demonstration of using Flipt + Dex as an OIDC provider.

## Requirements

- Docker
- Docker Compose

## Running

1. You're going to need an entry for `dex` in your `/etc/hosts` for this configuration.

This is so that we can support setups such as "Docker for Mac", which don't have full "host" network mode for Docker.

Add the following to your `/etc/hosts` file:

```
127.0.0.1 dex
```

e.g. `sudo sh -c 'echo "127.0.0.1 dex >> /etc/hosts"'`.

2. Run `docker-compose up` from this relative directory.
3. Navigate to [Flipt in your browser](http://localhost:8080/auth/v1/method/oidc/dex/authorize).
4. Click the `authorizeURL` link in the returned JSON blob.
5. Login using email: `admin@example.com` and password: `password`.
6. Select `Grant Access`.

From here you should be navigated back to Flipt and an authenticated session should be established.
