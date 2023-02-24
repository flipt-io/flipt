# NextJS Example

This is an example of how to use Flipt with a NextJS application.

It uses the [Flipt TypeScript SDK](https://github.com/flipt-io/flipt-node) to evalute feature flags from the Flipt API in two different ways:

1. Using the `useFlipt` hook to evalute the flag in the browser/client side (using the `useEffect` hook)
2. Using the [`getServerSideProps`](https://nextjs.org/docs/basic-features/data-fetching/get-server-side-props) function to evaluate the flags on the server side before rendering the page 

## Example

In this example, we are leveraging Flipt to prototype some personalization for our NextJS application. We want to show a different greeting message to the user at random, but we aren't sure if we should use client-side data fetching or server-side data fetching, so we are going to try both ways. We will use the Flipt TypeScript SDK to integrate with Flipt.

## Architecture

### Client Side

<p align="center">
    <img src="../images/nextjs-client-side.png" alt="Client Side Architecture" height=200 />
</p>

For the client-side example, we are using the `useFlipt` hook to evaluate the flag in the browser/client side (using the `useEffect` hook). The `useFlipt` hook returns an instance of our `FliptApiClient` object which is used to evalute the `language` flag, we can then use this value to render the greeting message.

Because we don't want to potentially expose any sensitive information, such as a Flipt API key, we are actually proxying the `/api/v1/evaluate` request through a reverse proxy via Caddy. This allows us to use the `FliptApiClient` directly in the NextJS application without exposing the API key. This has the additional benefit of not requiring us to publicly expose the Flipt server, and also allows us to only allow the `evaluate` request to be routed to Flipt. The Caddy configuration is defined in the `Caddyfile` file.

### Server Side

<p align="center">
    <img src="../images/nextjs-server-side.png" alt="Server Side Architecture" height=200 />
</p>

For the server-side example, we are using the [`getServerSideProps`](https://nextjs.org/docs/basic-features/data-fetching/get-server-side-props) function to evaluate the flags on the server side before rendering the page. This uses the `FliptApiClient` directly to evalute the `language` flag to then render the greeting message.

Since we are using server side rendering, we don't need to proxy the `/api/v1/evaluate` request through Caddy, we can just use the `FliptApiClient` directly in the NextJS application.

## Requirements

To run this example application you'll need:

* [Docker](https://docs.docker.com/install/)
* [docker-compose](https://docs.docker.com/compose/install/)

## Running the Example

1. Run `docker-compose up` from this directory
1. Open the Flipt UI (default: [http://localhost:8080](http://localhost:8080)) to browse the example Flags/Variants/Segments/etc that are pre-populated.
1. Open the NextJS UI (default: [http://localhost:3000](http://localhost:3000)) to see the example application.
1. Refresh the page to see the greeting messages change.
