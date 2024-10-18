# NextJS Examples

<p align="center">
    <img src="../images/nextjs.png" alt="NextJS Example" width=800 />
</p>

This repository contains examples of how to use Flipt with NextJS.

Both examples are using the same Flipt flag and are evaluating the flag within the application. 

Both examples use two different SDKs to evaluate the flag depending on where the evaluation is being performed.

- [Flipt Node SDK](https://www.npmjs.com/package/@flipt-io/flipt)
- [Flipt React SDK](https://www.npmjs.com/package/@flipt-io/flipt-client-react)

## Examples

Because NextJS supports two different routing strategies, there are two different examples:

- [Pages Router](./pages-router/README.md) (pre-NextJS 13)
- [App Router](./app-router/README.md) (NextJS 13+)

## Requirements

To run the examples, you'll need:

- [Docker](https://docs.docker.com/install/)
- [docker-compose](https://docs.docker.com/compose/install/)