# Installation

## Docker

The simplest way to run Flipt is via Docker. This streamlines the installation and configuration by using a reliable runtime.

### Prerequisites

Docker installation is required on the host, see the [official installation docs](https://docs.docker.com/install/).

!!! note
    Using a native Docker install instead of Docker Toolbox is recommended in order to use persisted volumes.

### Run the image

```shell
docker run -d \
    -p 8080:8080 \
    -p 9000:9000 \
    -v $HOME/flipt:/var/opt/flipt \
    markphelps/flipt:latest
```

This will download the image and start a Flipt container and publish ports needed to access the UI and backend server. All persistent Flipt data will be stored in `$HOME/flipt`.

!!! note
    `$HOME/flipt` is just used as an example, you can use any directory you would like on the host.

The Flipt container uses host mounted volumes to persist data:

| Host location | Container location | Purpose |
|---|---|---|
| $HOME/flipt  | /var/opt/flipt | For storing application data |

This allows data to persist between Docker container restarts.

!!! warning
    If you don't use mounted volumes to persist your data, your data will be lost when the container exits!

After starting the container you can visit [http://0.0.0.0:8080](http://0.0.0.0:8080) to view the application.

## Linux Packages

**Flipt RPM/DEB binary packages coming soon!**

## Download from GitHub

You can always download the latest release archive for your architecture from the [Releases](https://github.com/markphelps/flipt/releases) section on GitHub.

This archive contains the Flipt binary, configuration, database migrations, README, LICENSE and CHANGELOG files.

Copy the binary, config file and migrations to an accessible location on your host.

!!! note
    You will need to update the config file: `default.yml` if your migrations and database locations differ from the standard locations.

Run the Flipt binary with:

```shell
./flipt --config PATH_TO_YOUR_CONFIG
```

See the [Configuration](configuration.md) section for more details.
