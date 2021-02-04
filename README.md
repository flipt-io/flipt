<p align=center>
	<img src="logo.svg" alt="Flipt" width=200 height=200 />
</p>

<p align="center">An open-source, on-prem feature flag solution</p>

<hr />

![Flipt](demo.gif)

<div align="center">
    <a href="https://github.com/markphelps/flipt/actions">
        <img src="https://github.com/markphelps/flipt/workflows/Tests/badge.svg" alt="Build Status" />
    </a>
    <a href="https://bestpractices.coreinfrastructure.org/projects/3498">
	<img src="https://bestpractices.coreinfrastructure.org/projects/3498/badge">
    </a>
    <a href="https://codecov.io/gh/markphelps/flipt">
        <img src="https://codecov.io/gh/markphelps/flipt/branch/master/graph/badge.svg" alt="Coverage" />
    </a>
    <a href="https://goreportcard.com/report/github.com/markphelps/flipt">
        <img src="https://goreportcard.com/badge/github.com/markphelps/flipt" alt="Go Report Card" />
    </a>
    <a href="https://github.com/markphelps/flipt/releases">
        <img src="https://img.shields.io/github/release/markphelps/flipt.svg?style=flat" alt="Releases" />
    </a>
    <a href="https://hub.docker.com/r/markphelps/flipt">
        <img src="https://img.shields.io/docker/pulls/markphelps/flipt.svg" alt="Docker Pulls" />
    </a>
    <a href="https://github.com/avelino/awesome-go">
        <img src="https://awesome.re/mentioned-badge.svg" alt="Mentioned in Awesome Go" />
    </a>
</div>

<div align="center">
    <h4>
        <a href="https://flipt.io/docs/getting_started/">Documentation</a> |
        <a href="#features">Features</a> |
        <a href="#values">Values</a> |
        <a href="#examples">Examples</a> |
        <a href="#enterprise">Enterprise</a>
    </h4>
</div>

Flipt is an open source, on-prem feature flag application that allows you to run experiments across services in **your** environment.

Flipt can be deployed within your existing infrastructure so that you don't have to worry about your information being sent to a third party or the latency required to communicate across the internet.

Flipt supports use cases such as:

* Simple on/off feature flags to toggle functionality in your applications
* Rolling out features to a percentage of your customers
* Using advanced segmentation to target and serve users based on custom properties that you define

## Features

* Fast. Written in Go. Optimized for performance
* Stand alone, easy to run and configure
* Ability to create advanced distribution rules to target segments of users
* Native [GRPC](https://grpc.io/) client SDKs to integrate with your applications
* Simple REST API
* Modern UI and debug console
* Support for multiple databases (Postgres, MySQL, SQLite)
* Data import and export to allow storing your flags as code

## Values

* :lock: **Security** - HTTPS support. No data leaves your servers and you don't have to open your systems to the outside world to communicate with Flipt. It all runs within your existing infrastructure.
* :rocket: **Speed** - Since Flipt is co-located with your existing services, you do not have to communicate across the internet which can add excessive latency and slow down your applications.
* :white_check_mark: **Simplicity** - Flipt is a single binary with no external dependencies by default.
* :no_entry: **Privacy** - No telemetry data is collected or sent by Flipt. Ever.
* :thumbsup: **Compatability** - REST, GRPC, MySQL, Postgres, SQLite.. Flipt supports it all.

## Examples

Check out the [examples](/examples) to see how Flipt works.

Here's a [basic one](https://github.com/markphelps/flipt/tree/master/examples/basic) to get started!

## Try It

![Flipt Docker](cli.gif)

Try Flipt out yourself with Docker:

```bash
❯ docker run --rm -p 8080:8080 -p 9000:9000 -t markphelps/flipt:latest
```

Flipt UI will now be reachable at [http://localhost:8080/](http://localhost:8080).

For more permanent methods of running Flipt, see the [Installation](https://flipt.io/docs/installation/) section.

## GRPC Clients

* [Go](https://github.com/markphelps/flipt-grpc-go)
* [Ruby](https://github.com/markphelps/flipt-grpc-ruby)

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

## Enterprise

Need more features or support using Flipt within your Enterprise?

Please help me prioritize an Enterprise version of Flipt by filling out this [short survey](https://forms.gle/a4UBnv8LADYirA4c9)!

### Potential Features

* Business-friendly Licensing
* User Management and Audit Trail
* Multiple Environments (ex: dev/staging/prod)

## Author

| [![twitter/mark_a_phelps](https://secure.gravatar.com/avatar/274e2d4b1bbb9f86b454aebabad2cba1)](https://twitter.com/intent/user?screen_name=mark_a_phelps "Follow @mark_a_phelps on Twitter") |
|---|
| [Mark Phelps](https://markphelps.me/) |

## Contributing

I would love your help! Before submitting a PR, please read over the [Contributing](.github/contributing.md) guide.

No contribution is too small, whether it be bug reports/fixes, feature requests, documentation updates, or anything else that can help drive the project forward.

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="http://aaronraff.github.io"><img src="https://avatars0.githubusercontent.com/u/16910064?v=4" width="100px;" alt=""/><br /><sub><b>Aaron Raff</b></sub></a><br /><a href="https://github.com/markphelps/flipt/commits?author=aaronraff" title="Code">💻</a></td>
    <td align="center"><a href="http://twitter.com/rochacon"><img src="https://avatars2.githubusercontent.com/u/321351?v=4" width="100px;" alt=""/><br /><sub><b>Rodrigo Chacon</b></sub></a><br /><a href="https://github.com/markphelps/flipt/commits?author=rochacon" title="Code">💻</a></td>
    <td align="center"><a href="http://christopherdiehl.github.io"><img src="https://avatars0.githubusercontent.com/u/10383665?v=4" width="100px;" alt=""/><br /><sub><b>Christopher Diehl</b></sub></a><br /><a href="https://github.com/markphelps/flipt/commits?author=christopherdiehl" title="Code">💻</a></td>
    <td align="center"><a href="https://www.andrewzallen.com"><img src="https://avatars3.githubusercontent.com/u/37206?v=4" width="100px;" alt=""/><br /><sub><b>Andrew Z Allen</b></sub></a><br /><a href="https://github.com/markphelps/flipt/commits?author=achew22" title="Documentation">📖</a></td>
    <td align="center"><a href="http://sf.khepin.com"><img src="https://avatars3.githubusercontent.com/u/455656?v=4" width="100px;" alt=""/><br /><sub><b>Sebastien Armand</b></sub></a><br /><a href="https://github.com/markphelps/flipt/commits?author=khepin" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/badboyd"><img src="https://avatars0.githubusercontent.com/u/20040686?v=4" width="100px;" alt=""/><br /><sub><b>Dat Tran</b></sub></a><br /><a href="https://github.com/markphelps/flipt/commits?author=badboyd" title="Code">💻</a></td>
    <td align="center"><a href="http://twitter.com/jon_perl"><img src="https://avatars2.githubusercontent.com/u/1136652?v=4" width="100px;" alt=""/><br /><sub><b>Jon Perl</b></sub></a><br /><a href="https://github.com/markphelps/flipt/commits?author=jperl" title="Tests">⚠️</a> <a href="https://github.com/markphelps/flipt/commits?author=jperl" title="Code">💻</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://or-e.net"><img src="https://avatars1.githubusercontent.com/u/2883824?v=4" width="100px;" alt=""/><br /><sub><b>Or Elimelech</b></sub></a><br /><a href="https://github.com/markphelps/flipt/commits?author=vic3lord" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/giddel"><img src="https://avatars0.githubusercontent.com/u/10463018?v=4" width="100px;" alt=""/><br /><sub><b>giddel</b></sub></a><br /><a href="https://github.com/markphelps/flipt/commits?author=giddel" title="Code">💻</a></td>
  </tr>
</table>

<!-- markdownlint-enable -->
<!-- prettier-ignore-end -->
<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!
