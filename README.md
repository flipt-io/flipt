<img referrerpolicy="no-referrer-when-downgrade" src="https://static.scarf.sh/a.png?x-pxid=dc285028-90ba-436b-9a4a-e0d826a2c986" alt="" />

<p align=center>
    <img src="logo.svg" alt="Flipt" width=275 height=96 />
</p>

<p align="center">An enterprise-ready, GitOps and CloudNative, feature management solution</p>

<hr />

<p align="center">
    <img src=".github/images/dashboard.png" alt="Flipt Dashboard" width=600 />
</p>

<br clear="both"/>

<div align="center">
    <a href="https://github.com/flipt-io/flipt/releases">
        <img src="https://img.shields.io/github/release/flipt-io/flipt.svg?style=flat" alt="Releases" />
    </a>
    <a href="https://github.com/flipt-io/flipt/blob/main/LICENSE">
        <img src="https://img.shields.io/github/license/flipt-io/flipt.svg" alt="GPL 3.0" />
    </a>
    <a href="https://codecov.io/gh/flipt-io/flipt">
        <img src="https://codecov.io/gh/flipt-io/flipt/branch/main/graph/badge.svg" alt="Coverage" />
    </a>
    <a href="https://goreportcard.com/report/github.com/flipt-io/flipt">
        <img src="https://goreportcard.com/badge/github.com/flipt-io/flipt" alt="Go Report Card" />
    </a>
    <a href="https://github.com/avelino/awesome-go">
        <img src="https://awesome.re/mentioned-badge.svg" alt="Mentioned in Awesome Go" />
    </a>
    <a href="https://flipt.io/discord">
        <img alt="Discord" src="https://img.shields.io/discord/960634591000014878?color=%238440f1&label=Discord&logo=discord&logoColor=%238440f1&style=flat" />
    </a>
</div>

<div align="center">
    <h4>
        <a href="https://www.flipt.io/docs/introduction">Docs</a> •
        <a href="http://www.flipt.io">Website</a> •
        <a href="http://blog.flipt.io">Blog</a> •
        <a href="https://community.flipt.io/">Feedback</a> •
        <a href="#contributing">Contributing</a> •
        <a href="https://www.flipt.io/discord">Discord</a>
    </h4>
</div>

> [!IMPORTANT]  
> This branch is a work in progress for a v2 version of Flipt. This is not a stable release and should not be used in production. The v2 branch is a major refactor of the codebase with the goal of support Git and object storage as the primary storage backends. See [PLAN.md](PLAN.md) for more information.

<br clear="both"/>

[Flipt](https://www.flipt.io) enables you to follow DevOps best practices and separate releases from deployments. Built with high-performance engineering organizations in mind.

Flipt can be deployed within your existing infrastructure so that you don't have to worry about your information being sent to a third party or the latency required to communicate across the internet.

With our [GitOps-friendly functionality](https://www.flipt.io/docs/guides/get-going-with-gitops), you can easily integrate Flipt into your CI/CD workflows to enable continuous configuration and deployment with confidence.

<br clear="both"/>

<p align="center">
    <a href="https://www.producthunt.com/posts/flipt-cloud?embed=true&utm_source=badge-featured&utm_medium=badge&utm_souce=badge-flipt&#0045;cloud" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=498373&theme=light" alt="Flipt&#0032;Cloud - Feature&#0032;flags&#0046;&#0032;Powered&#0032;by&#0032;Git | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>
    <a href="https://devhunt.org/tool/flipt" title="DevHunt - Tool of the Week"><img src="./.github/images/devhunt-badge.png" width=225 alt="DevHunt - Tool of the Week" /></a>&nbsp;
    <a href="https://console.dev/tools/flipt" title="Visit Console - the best tools for developers"><img src="./.github/images/console-badge.png" width=250 alt="Console - Developer Tool of the Week" /></a>
</p>

## Use cases

Flipt supports use cases such as:

- Enabling [trunk-based development](https://trunkbaseddevelopment.com/) workflows
- Testing new features internally during development before releasing them fully in production
- Ensuring overall system safety by guarding new releases with an emergency kill switch
- Gating certain features for different permission levels allows you to control who sees what
- Enabling continuous configuration by changing values during runtime without additional deployments

<br clear="both"/>

## Values

- 🔒 **Security** - HTTPS, OIDC, JWT, OAuth, K8s Service Token, and API Token authentication methods supported out of the box.
- 🎛️ **Control** - No data leaves your servers and you don't have to open your systems to the outside world to communicate with Flipt. It all runs within your existing infrastructure.
- 🚀 **Speed** - Since Flipt is co-located with your existing services, you do not have to communicate across the internet which can add excessive latency and slow down your applications.
- ✅ **Simplicity** - Flipt is a single binary with no external dependencies by default.
- 👍 **Compatibility** - GRPC, REST, MySQL, Postgres, SQLite, Redis, ClickHouse, Prometheus, OpenTelemetry, and more.

<br clear="both"/>

## Features

- Stand-alone, single binary that's easy to run and [configure](https://www.flipt.io/docs/configuration/overview)
- Ability to create advanced distribution rules to target segments of users
- Modern UI and debug console with dark mode 🌙
- Works with [Prometheus](https://prometheus.io/) and [OpenTelemetry](https://opentelemetry.io/) out of the box 🔋
- CloudNative [Filesystem, Object, Git, and OCI declarative storage backends](https://www.flipt.io/docs/configuration/storage#declarative) to support GitOps workflows and more.
- Audit logging with Webhook support to track changes to your data

Are we missing a feature that you'd like to see? [Let us know by opening an issue!](https://github.com/flipt-io/flipt/issues)

<br clear="both"/>

## Contributing

We would love your help! Before submitting a PR, please read over the [Contributing](CONTRIBUTING.md) guide.

No contribution is too small, whether it be bug reports/fixes, feature requests, documentation updates, or anything else that can help drive the project forward.

Not sure how to get started? You can:

- [Book a pairing session/code walkthrough](https://calendly.com/flipt-mark/30) with one of our teammates!
- Join our [Discord](https://www.flipt.io/discord), and ask any questions there

- Dive into any of the open issues, here are some examples: 
  - [Good First Issues](https://github.com/flipt-io/flipt/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)
  - [Backend](https://github.com/flipt-io/flipt/issues?q=is%3Aissue+is%3Aopen+label%3Ago)
  - [Frontend](https://github.com/flipt-io/flipt/issues?q=is%3Aopen+is%3Aissue+label%3Aui)

Review the [Architecture](ARCHITECTURE.md) and [Development](DEVELOPMENT.md) documentation for more information on how Flipt works.

<br clear="both"/>

## Community

For help and discussion around Flipt, feature flag best practices, and more, join us on [Discord](https://www.flipt.io/discord).

<br clear="both"/>

<!--## Try It

Get started in seconds. Try the latest version of Flipt for yourself.

### Local

```shell
curl -fsSL https://get.flipt.io/install | sh
```

### Deploy 

<div>
    <a href="https://marketplace.digitalocean.com/apps/flipt" alt="Deploy to DigitalOcean">
        <img width="200" alt="Deploy to DigitalOcean" src="https://www.deploytodo.com/do-btn-blue.svg"/>
    </a>&nbsp;
    <a href="https://render.com/deploy" alt="Deploy to Render">
        <img width="150" alt="Deploy to Render" src="http://render.com/images/deploy-to-render-button.svg" />
    </a>&nbsp;
    <a href="https://railway.app/template/dz-JCO" alt="Deploy to Railway">
      <img width="150" alt="Deploy to Railway" src="https://railway.app/button.svg" />
    </a>
    <a href="https://app.koyeb.com/deploy?type=docker&image=docker.flipt.io/flipt/flipt&ports=8080;http;/&name=flipt-demo" alt="Deploy to Koyeb">
      <img width="150" alt="Deploy to Koyeb" src="https://www.koyeb.com/static/images/deploy/button.svg" />
    </a>
</div>

### Sandbox

[Try Flipt](https://try.flipt.io) in a deployed environment!

**Note:** The database gets cleared **every 30 minutes** in this sandbox environment!

### Homebrew :beer:

```bash
brew install flipt-io/brew/flipt
brew services start flipt

# or run in the foreground
flipt
```

Flipt UI will now be reachable at [http://127.0.0.1:8080/](http://127.0.0.1:8080).

### Docker :whale:

```bash
docker run --rm -p 8080:8080 -p 9000:9000 -t docker.flipt.io/flipt/flipt:latest
```

Flipt UI will now be reachable at [http://127.0.0.1:8080/](http://127.0.0.1:8080).

For more permanent methods of running Flipt, see the [Installation](https://flipt.io/docs/installation/) section.

### Nightly Build

Like to live on the edge? Can't wait for the next release? Our nightly builds include the latest changes on `main` and are built.. well.. nightly.

```bash
docker run --rm -p 8080:8080 -p 9000:9000 -t docker.flipt.io/flipt/flipt:nightly
``` 
-->

<br clear="both"/>

## Supports

<p align="center">
    <img src="./logos/redis.svg" alt="Redis" height=75 />
    <img src="./logos/prometheus.svg" alt="Prometheus" height=75 />
    <img src="./logos/openid.svg" alt="OpenID" height=75 />
    <img src="./logos/opentelemetry.svg" alt="OpenTelemetry" height=75 />
    <img src="./logos/git.svg" alt="Git" height=50 />
</p>

<br clear="both"/>

## Integration

Check out our [integration documentation](https://flipt.io/docs/integration/) for more info on how to integrate Flipt into your existing applications.

There are two ways to evaluate feature flags with Flipt:

- [Server Side](#server-side-evaluation)
- [Client Side](#client-side-evaluation)

### Server Side Evaluation

Server-side evaluation is the most common way to evaluate feature flags. This is where your application makes a request to Flipt to evaluate a feature flag and Flipt responds with the result of the evaluation.

Flipt exposes two different APIs for performing server-side evaluation:

- [GRPC](#grpc)
- [REST](#rest)

#### GRPC

Flipt is equipped with a fully functional GRPC API. GRPC is a high-performance, low-latency, binary protocol that is used by many large-scale companies such as Google, Netflix, and more.

See our [GRPC Server SDK documentation](https://www.flipt.io/docs/integration/server/grpc) for the latest information.

#### REST

Flipt is equipped with a fully functional REST API. The Flipt UI is completely backed by this same API. This means that anything that can be done in the Flipt UI can also be done via the REST API.

The [Flipt REST API](https://www.flipt.io/docs/reference/overview) can also be used with any language that can make HTTP requests.

See our [REST Server SDK documentation](https://www.flipt.io/docs/integration/server/rest) for the latest information.

### Client Side Evaluation

Client-side evaluation is a great way to reduce the number of requests that your application needs to make to Flipt. This is done by retrieving all of the feature flags that your application needs to evaluate and then evaluating them locally.

See our [Client SDK documentation](https://www.flipt.io/docs/integration/client) for the latest information.

<br clear="both"/>

## Release Cadence

Flipt follows [semantic versioning](https://semver.org/) for versioning.

We aim to release a new minor version of Flipt every 2-3 weeks. This allows us to quickly iterate on new features.
Bug fixes and security patches (patch versions) will be released as needed.

<br clear="both"/>

## Development

[Development](DEVELOPMENT.md) documentation is available for those interested in contributing to Flipt.

We welcome contributions of any kind, including but not limited to bug fixes, feature requests, documentation improvements, and more. Just open an issue or pull request and we'll be happy to help out!

<br clear="both"/>

[![Open in Codespaces](https://github.com/codespaces/badge.svg)](https://github.com/codespaces/new/?repo=flipt-io/flipt)

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/flipt-io/flipt)

<br clear="both"/>

## Examples

Check out the [examples](/examples) to see how Flipt works in different use cases.

<br clear="both"/>

## Licensing

There are currently two types of licenses in place for Flipt:

1. Client License
2. Server License

### Client License

All of the code required to generate GRPC clients in other languages as well as the [Go SDK](/sdk/go) are licensed under the [MIT License](https://spdx.org/licenses/MIT.html).

This code exists in the [rpc/](rpc/) directory.

The client code is the code that you would integrate into your applications, which is why a more permissive license is used.

### Server License

The server code is licensed under the [Fair Core License, Version 1.0, MIT Future License](https://github.com/flipt-io/flipt/blob/main/LICENSE).

See [fcl.dev](https://fcl.dev) for more information.
