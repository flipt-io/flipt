# Load Testing

This is a load testing script for Flipt. It uses the [k6](https://k6.io/) tool to simulate load on the Flipt server.

## Prerequisites

- [k6](https://k6.io/)
- [Flipt](https://flipt.io/)

## Steps

1. Install k6
2. Install or build Flipt
3. Import data into Flipt
4. Run the load test

## Import data into Flipt

```bash
flipt import import.yaml
```

## Run the load test

```bash
FLIPT_ADDR=http://localhost:8080 k6 run loadtest.js
```

To run with the k6 Web UI:

```bash
FLIPT_ADDR=http://localhost:8080 K6_WEB_DASHBOARD=true k6 run loadtest.js
```
