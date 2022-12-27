# Flipt UI

This repository contains the new (beta) UI for [Flipt](https://github.com/flipt-io/flipt).

It is built with [Next.js](https://nextjs.org/) and [Tailwind CSS](https://tailwindcss.com/) and is currently in beta and under active development.

## Development

### Prerequisites

- [Node.js](https://nodejs.org/en/) (v16 or later)
- [Docker](https://www.docker.com/products/docker-desktop) (optional but recommended)

### Flipt API

You'll need to run the Flipt API backend locally to develop the UI.

The simplest way to do this is to use the [Flipt Docker image](https://hub.docker.com/r/flipt/flipt). See the [documentation](https://www.flipt.io/docs/installation#run-the-image) for more details.

Note: It's recommended to use the `latest` tag for the Docker image if you want the latest stable release, otherwise you can use the `nightly` tag for the most up to date code changes.

```bash
docker run -d \
    -p 8080:8080 \
    -p 9000:9000 \
    -v $HOME/flipt:/var/opt/flipt \
    flipt/flipt:nightly
```

### Environment Variables

The UI use environment variables to know how to connect to the Flipt API. You can set these in an `.env.local` file in the root of the project.

There's an example file in the root of the project called `.env.local.example`. You can copy this to `.env.local` and use or modify the values.

Namely, you'll need to ensure to set the `NEXT_PUBLIC_BASE_URL` variable to the URL of the Flipt API.

```bash
export NEXT_PUBLIC_BASE_URL=http://localhost:8080
```

### Running the UI

Run the development server:

```bash
npm run dev
# or
yarn dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.
