<p align=center>
    <img src="logo.svg" alt="Flipt" width=275 height=96 />
</p>

<p align="center">An open source, self-hosted feature flag solution</p>

<hr />

<p align="center">
    <img src=".github/images/flags.png" alt="Flipt Dashboard" width=600 />
</p>

<div align="center">
    <a href="https://github.com/flipt-io/flipt/releases">
        <img src="https://img.shields.io/github/release/flipt-io/flipt.svg?style=flat" alt="Releases" />
    </a>
    <a href="https://github.com/flipt-io/flipt/actions">
        <img src="https://github.com/flipt-io/flipt/workflows/Tests/badge.svg" alt="Build Status" />
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
    <a href="https://bestpractices.coreinfrastructure.org/projects/3498">
        <img src="https://bestpractices.coreinfrastructure.org/projects/3498/badge">
    </a>
    <a href="https://github.com/avelino/awesome-go">
        <img src="https://awesome.re/mentioned-badge.svg" alt="Mentioned in Awesome Go" />
    </a>
    <a href="https://magefile.org">
        <img src="https://magefile.org/badge.svg" alt="Built with Mage" />
    </a>
    <a href="https://discord.gg/kRhEqG2TEZ">
        <img alt="Discord" src="https://img.shields.io/discord/960634591000014878?color=%238440f1&label=Discord&logo=discord&logoColor=%238440f1&style=flat">
    </a>
    <a href="https://volta.net/embed/eyJzdGF0dXNlcyI6WyJ0cmlhZ2UiLCJiYWNrbG9nIiwidG9kbyIsImluX3Byb2dyZXNzIiwiaW5fcmV2aWV3IiwiZG9uZSIsInJlbGVhc2VkIiwiY2FuY2VsbGVkIl0sImZpbHRlcnMiOnt9LCJvd25lciI6ImZsaXB0LWlvIiwibmFtZSI6ImZsaXB0In0=">
        <img alt="Public Roadmap" src="https://img.shields.io/badge/Volta-Public%20Roadmap-black">
    </a>
</div>

<div align="center">
    <h4>
        <a href="https://www.flipt.io/docs/introduction">Documentation</a> •
        <a href="#usecases">Usecases</a> •
        <a href="#features">Features</a> •
        <a href="#values">Values</a> •
        <a href="#examples">Examples</a> •
        <a href="#integration">Integration</a> •
        <a href="#community">Community</a>
    </h4>
</div>

Flipt is an open-source, self-hosted feature flag application that allows you to run experiments across services in **your** environment.

Flipt can be deployed within your existing infrastructure so that you don't have to worry about your information being sent to a third party or the latency required to communicate across the internet.

<br clear="both"/>

## Usecases

Flipt supports use cases such as:

- Enabling [trunk-based development](https://trunkbaseddevelopment.com/) workflows
- Testing new features internally during development before releasing them fully in production
- Ensuring overall system safety by guarding new releases with an emergency kill switch
- Gating certain features for different permission levels allowing you to control who sees what
- Enabling continuous configuration by changing values during runtime without additional deployments

<br clear="both"/>

## Features

<img align="right" src=".github/images/console.png" alt="Flipt Console" width=40% />

- Fast. Written in Go. Optimized for performance
- Stand alone, easy to run and configure
- Ability to create advanced distribution rules to target segments of users
- Native [GRPC](https://grpc.io/) client SDKs to integrate with your existing applications easily
- Powerful REST API
- Modern, mobile-friendly 📱 UI and debug console
- Support for multiple databases (Postgres, MySQL, SQLite, CockroachDB)
- Data import and export to allow storing your data as code
- Cloud-ready :cloud:. Runs anywhere: bare metal, PaaS, K8s, with Docker or without.

<br clear="both"/>

## Values

- :lock: **Security** - HTTPS support. [OIDC](https://www.flipt.io/docs/authentication/methods#openid-connect) and [Static Token](https://www.flipt.io/docs/authentication/methods#static-token) authentication. No data leaves your servers and you don't have to open your systems to the outside world to communicate with Flipt. It all runs within your existing infrastructure.
- :rocket: **Speed** - Since Flipt is co-located with your existing services, you do not have to communicate across the internet which can add excessive latency and slow down your applications.
- :white_check_mark: **Simplicity** - Flipt is a single binary with no external dependencies by default.
- :thumbsup: **Compatibility** - REST, GRPC, MySQL, Postgres, CockroachDB, SQLite, Redis... Flipt supports it all.
- :eyes: **Observability** - Flipt integrates with [Prometheus](https://prometheus.io/) and [OpenTelemetry](https://opentelemetry.io/) to provide metrics and tracing. We support sending trace data to [Jaeger](https://www.jaegertracing.io/), [Zipkin](https://zipkin.io/), and [OpenTelemetry Protocol (OTLP)](https://opentelemetry.io/docs/reference/specification/protocol/) backends.

<br clear="both"/>

## Works With

<p align="center">
    <img src="./logos/sqlite.svg" alt="SQLite" width=150 height=150 />
    <img src="./logos/mysql.svg" alt="MySQL" width=150 height=150 />
    <img src="./logos/postgresql.svg" alt="PostgreSQL" width=150 height=150 />
    <img src="./logos/cockroachdb.svg" alt="CockroachDB" width=100 height=150 />
</p>
<p align="center">
    <img src="./logos/redis.svg" alt="Redis" width=150 height=150 />
    <img src="./logos/prometheus.svg" alt="Prometheus" width=150 height=150 />
    <img src="./logos/openid.svg" alt="OpenID" width=125 height=125 />
    <img src="./logos/opentelemetry.svg" alt="OpenTelemetry" width=150 height=150 />
</p>

## Try It

Try the latest version of Flipt for yourself.

### Sandbox

[Try Flipt](https://try.flipt.io) in a deployed environment!

**Note:** The database gets cleared **every 30 minutes** in this sandbox environment!

### Docker

```bash
docker run --rm -p 8080:8080 -p 9000:9000 -t flipt/flipt:latest
```

Flipt UI will now be reachable at [http://127.0.0.1:8080/](http://127.0.0.1:8080).

For more permanent methods of running Flipt, see the [Installation](https://flipt.io/docs/installation/) section.

### Nightly Build

Like to live on the edge? Can't wait for the next release? Our nightly builds include the latest changes on `main` and are built.. well.. nightly.

```bash
docker run --rm -p 8080:8080 -p 9000:9000 -t flipt/flipt:nightly
```

<br clear="both"/>

## Examples

Check out the [examples](/examples) to see how Flipt works in different use cases.

<br clear="both"/>

## Integration

Check out the [integration docs](https://flipt.io/docs/integration/) for more info on how to integrate Flipt into your existing applications.

### REST API

Flipt is equipped with a fully functional REST API. In fact, the Flipt UI is completely backed by this same API. This means that anything that can be done in the Flipt UI can also be done via the REST API.

The [Flipt REST API](https://www.flipt.io/docs/reference/overview) can also be used with any language that can make HTTP requests.

### Official SDK

Our SDK is designed to support either HTTP or gRPC (configurable based on your needs).

- [Go](./sdk/go)

### Official REST Client Libraries

- [Node/TypeScript](https://github.com/flipt-io/flipt-node)
- [Java](https://github.com/flipt-io/flipt-java)
- [Rust](https://github.com/flipt-io/flipt-rust)

:exclamation: Offical REST clients in more languages coming soon.

### Official GRPC Client Libraries

- [Go](https://github.com/flipt-io/flipt-grpc-go)
- [Ruby](https://github.com/flipt-io/flipt-grpc-ruby)

:exclamation: Offical GRPC clients in more languages coming soon.

### Third-Party Client Libraries

Client libraries built by awesome people from the Open Source community.

Note: These libraries are not maintained by the Flipt team and may not be up to date with the latest version of Flipt. Please open an issue or pull request on the library’s repository if you find any issues.

| Library                                                             | Language   | Author                                                   | Desc                                                                                            |
| ------------------------------------------------------------------- | ---------- | -------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| [flipt-grpc-python](https://github.com/getsentry/flipt-grpc-python) | Python     | [@getsentry](https://github.com/getsentry)               | Python GRPC bindings for Flipt                                                                  |
| [rflipt](https://github.com/christopherdiehl/rflipt)                | React      | [@christopherdiehl](https://github.com/christopherdiehl) | Components/example project to control React features backed by Flipt                            |
| [flipt-php](https://github.com/fetzi/flipt-php)                     | PHP        | [@fetzi](https://github.com/fetzi)                       | Package for evaluating feature flags via the Flipt REST API using [HTTPlug](http://httplug.io/) |
| [flipt-js](https://github.com/betrybe/flipt-js)                     | Javascript | [@betrybe](https://github.com/betrybe)                   | Flipt library for JS that allows rendering components based on Feature Flags 🎉                 |

### Generate Your Own

If a client in your language is not available for download, you can easily generate one yourself using the existing [protobuf definition](https://github.com/flipt-io/flipt/blob/main/rpc/flipt/flipt.proto). The [GRPC documentation](https://grpc.io/docs/) has extensive examples of how to generate GRPC clients in each supported language.

<br clear="both"/>

## Licensing

There are currently two types of licenses in place for Flipt:

1. Client License
2. Server License

### Client License

All of the code required to generate GRPC clients in other languages as well as the existing GRPC Go client are licensed under the [MIT License](https://spdx.org/licenses/MIT.html).

This code exists in the [rpc/](rpc/) directory.

The client code is the code that you would integrate into your applications, which is why a more permissive license is used.

### Server License

The server code is licensed under the [GPL 3.0 License](https://spdx.org/licenses/GPL-3.0.html).

See [LICENSE](LICENSE).

<br clear="both"/>

## Logos

Some of the companies depending on Flipt in production.

<p align="center">
    <a href="https://paradigm.co">
        <img src="./logos/users/paradigm.png" alt="Paradigm" />
    </a>&nbsp;&nbsp;
    <a href="https://rokt.com">
        <img src="./logos/users/rokt.svg" alt="Rokt" width="200"/>
    </a>&nbsp;&nbsp;
    <a href="https://asphaltbot.com">
        <img src="./logos/users/asphaltlogo.png" alt="Asphalt" />
    </a>&nbsp;&nbsp;
    <a href="https://prose.com">
        <img src="./logos/users/prose.png" alt="Prose" width="200"/>
    </a>
</p>

Using Flipt at your company? Open a PR and add your logo here!

<br clear="both"/>

## Community

For help and discussion around Flipt, feature flag best practices, and more, join us on [Discord](https://www.flipt.io/discord).

<br clear="both"/>

## Feedback

If you are a user of Flipt we'd really :heart: it if you could leave a [testimonial](https://testimonial.to/flipt) on how Flipt is working for you.

<br clear="both"/>

## Contributing

We would love your help! Before submitting a PR, please read over the [Contributing](.github/contributing.md) guide.

No contribution is too small, whether it be bug reports/fixes, feature requests, documentation updates, or anything else that can help drive the project forward.

Check out our [public roadmap](https://volta.net/embed/eyJzdGF0dXNlcyI6WyJ0cmlhZ2UiLCJiYWNrbG9nIiwidG9kbyIsImluX3Byb2dyZXNzIiwiaW5fcmV2aWV3IiwiZG9uZSIsInJlbGVhc2VkIiwiY2FuY2VsbGVkIl0sImZpbHRlcnMiOnt9LCJvd25lciI6ImZsaXB0LWlvIiwibmFtZSI6ImZsaXB0In0=) to see what we're working on and where you can help.

<br clear="both"/>

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tbody>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="http://aaronraff.github.io"><img src="https://avatars0.githubusercontent.com/u/16910064?v=4?s=100" width="100px;" alt="Aaron Raff"/><br /><sub><b>Aaron Raff</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=aaronraff" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://twitter.com/rochacon"><img src="https://avatars2.githubusercontent.com/u/321351?v=4?s=100" width="100px;" alt="Rodrigo Chacon"/><br /><sub><b>Rodrigo Chacon</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=rochacon" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://christopherdiehl.github.io"><img src="https://avatars0.githubusercontent.com/u/10383665?v=4?s=100" width="100px;" alt="Christopher Diehl"/><br /><sub><b>Christopher Diehl</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=christopherdiehl" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://www.andrewzallen.com"><img src="https://avatars3.githubusercontent.com/u/37206?v=4?s=100" width="100px;" alt="Andrew Z Allen"/><br /><sub><b>Andrew Z Allen</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=achew22" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://sf.khepin.com"><img src="https://avatars3.githubusercontent.com/u/455656?v=4?s=100" width="100px;" alt="Sebastien Armand"/><br /><sub><b>Sebastien Armand</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=khepin" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/badboyd"><img src="https://avatars0.githubusercontent.com/u/20040686?v=4?s=100" width="100px;" alt="Dat Tran"/><br /><sub><b>Dat Tran</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=badboyd" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://twitter.com/jon_perl"><img src="https://avatars2.githubusercontent.com/u/1136652?v=4?s=100" width="100px;" alt="Jon Perl"/><br /><sub><b>Jon Perl</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=jperl" title="Tests">⚠️</a> <a href="https://github.com/flipt-io/flipt/commits?author=jperl" title="Code">💻</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://or-e.net"><img src="https://avatars1.githubusercontent.com/u/2883824?v=4?s=100" width="100px;" alt="Or Elimelech"/><br /><sub><b>Or Elimelech</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=vic3lord" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/giddel"><img src="https://avatars0.githubusercontent.com/u/10463018?v=4?s=100" width="100px;" alt="giddel"/><br /><sub><b>giddel</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=giddel" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://eduar.do"><img src="https://avatars.githubusercontent.com/u/959623?v=4?s=100" width="100px;" alt="Eduardo"/><br /><sub><b>Eduardo</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=edumucelli" title="Documentation">📖</a> <a href="https://github.com/flipt-io/flipt/commits?author=edumucelli" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/itaischwartz"><img src="https://avatars.githubusercontent.com/u/60180089?v=4?s=100" width="100px;" alt="Itai Schwartz"/><br /><sub><b>Itai Schwartz</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=itaischwartz" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://bandism.net/"><img src="https://avatars.githubusercontent.com/u/22633385?v=4?s=100" width="100px;" alt="Ikko Ashimine"/><br /><sub><b>Ikko Ashimine</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=eltociear" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://sagikazarmark.hu"><img src="https://avatars.githubusercontent.com/u/1226384?v=4?s=100" width="100px;" alt="Márk Sági-Kazár"/><br /><sub><b>Márk Sági-Kazár</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=sagikazarmark" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/pietdaniel"><img src="https://avatars.githubusercontent.com/u/1924983?v=4?s=100" width="100px;" alt="Dan Piet"/><br /><sub><b>Dan Piet</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=pietdaniel" title="Code">💻</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/amayvs"><img src="https://avatars.githubusercontent.com/u/842194?v=4?s=100" width="100px;" alt="Amay Shah"/><br /><sub><b>Amay Shah</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=amayvs" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/kevin-ip"><img src="https://avatars.githubusercontent.com/u/28875408?v=4?s=100" width="100px;" alt="kevin-ip"/><br /><sub><b>kevin-ip</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=kevin-ip" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/albertchae"><img src="https://avatars.githubusercontent.com/u/217050?v=4?s=100" width="100px;" alt="albertchae"/><br /><sub><b>albertchae</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=albertchae" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://thomas.sickert.dev"><img src="https://avatars.githubusercontent.com/u/11492877?v=4?s=100" width="100px;" alt="Thomas Sickert"/><br /><sub><b>Thomas Sickert</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=tsickert" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/jalaziz"><img src="https://avatars.githubusercontent.com/u/247849?v=4?s=100" width="100px;" alt="Jameel Al-Aziz"/><br /><sub><b>Jameel Al-Aziz</b></sub></a><br /><a href="#platform-jalaziz" title="Packaging/porting to new platform">📦</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://george.macro.re"><img src="https://avatars.githubusercontent.com/u/1253326?v=4?s=100" width="100px;" alt="George"/><br /><sub><b>George</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=GeorgeMac" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://syntaqx.com"><img src="https://avatars.githubusercontent.com/u/6037730?v=4?s=100" width="100px;" alt="Chase Pierce"/><br /><sub><b>Chase Pierce</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=syntaqx" title="Code">💻</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="http://showwin.asia"><img src="https://avatars.githubusercontent.com/u/1732016?v=4?s=100" width="100px;" alt="ITO Shogo"/><br /><sub><b>ITO Shogo</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=showwin" title="Tests">⚠️</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/yquansah"><img src="https://avatars.githubusercontent.com/u/13950726?v=4?s=100" width="100px;" alt="Yoofi Quansah"/><br /><sub><b>Yoofi Quansah</b></sub></a><br /><a href="https://github.com/flipt-io/flipt/commits?author=yquansah" title="Code">💻</a></td>
    </tr>
  </tbody>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!
