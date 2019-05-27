<p align=center>
	<img src="logo.svg" alt="Flipt" width=200 height=200 />
</p>

<p align="center">A feature flag solution that runs in your existing infrastructure</p>

<hr />

![Flipt](docs/assets/images/flipt.png)

[![Build Status](https://travis-ci.com/markphelps/flipt.svg?token=TBiDDmnBkCmRa867CqCG&branch=master)](https://travis-ci.com/markphelps/flipt)
[![Coverage Status](https://coveralls.io/repos/github/markphelps/flipt/badge.svg)](https://coveralls.io/github/markphelps/flipt)
[![Go Report Card](https://goreportcard.com/badge/github.com/markphelps/flipt)](https://goreportcard.com/report/github.com/markphelps/flipt)
[![GitHub Release](https://img.shields.io/github/release/markphelps/flipt.svg?style=flat)](https://github.com/markphelps/flipt/releases)
[![Gitter](https://badges.gitter.im/markphelps/flipt.svg)](https://gitter.im/markphelps/flipt?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

## Documentation

[https://flipt.dev/](https://flipt.dev/)

## What is Flipt

Flipt is an open source feature flag application that allows you to run experiments across services in **your** environment.

This means that you can deploy Flipt within your existing infrastructure and not have to worry about your information being sent to a third party or the latency required to communicate across the internet.

Flipt includes native client SDKs as well as a REST API so you can choose how to best integrate Flipt with your applications.

## :warning: Beta Software

Flipt is still considered beta software until the 1.0 release. This means that there are likely bugs and features/configuration may change between releases. Attempts will be made to maintain backwards compatibility whenever possible.

## Flipt Features

Flipt enables you to add [feature flag](https://martinfowler.com/bliki/FeatureToggle.html) support to your existing applications, with a simple, single UI and API.

This can range from simple on/off feature flags to more advanced use cases where you want to be able to rollout different versions of a feature to percentages of your users.

Flipt features include:

* Fast. Written in Go. Optimized for performance
* Stand alone, easy to run and configure
* Ability to create advanced distribution rules to target segments of users
* Native GRPC client SDKs to integrate with your applications
* Simple REST API
* Modern UI and debug console

## Why Flipt

Flipt allows you to focus on building your applications without having to worry about implementing your own feature flag solution that works across your infrastructure.

On top of this, Flipt provides a nice, modern UI so that you can always monitor the state of your feature flags and experiments in a single place.

## Running Flipt

Flipt is a single, self contained binary that you run on your own servers or cloud infrastructure. There are a multitude of benefits to running Flipt yourself, including:

* :lock: **Security** - No data leaves your servers and you don't have to open your systems to the outside world to communicate with Flipt. It all runs within your existing infrastructure.
* :rocket: **Speed** - Since Flipt is co-located with your existing services, you do not have to communicate across the internet to another application running on the other side of the world which can add excessive latency and slow down your applications.
* :white_check_mark: **Simplicity** - Flipt is a single binary with no external dependencies by default.

### Try It

```bash
❯ docker run --rm -p 8080:8080 -p 9000:9000 markphelps/flipt:latest
```

Flipt UI will now be reachable at [http://localhost:8080/](http://localhost:8080).

For more permanent methods of running Flipt, see the [Installation](https://flipt.dev/installation/) section.

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

If there are any concerns about the use of this license for the server, please [open an issue](https://github.com/markphelps/flipt/issues/new) on GitHub so that we can discuss publicly.

## Author

* Website: [https://markphelps.me](https://markphelps.me)
* Twitter: [@mark_a_phelps](https://twitter.com/mark_a_phelps)
* Email: _mark.aaron.phelps at gmail.com_

## Contributing

I would love your help! Before submitting a PR, please read over the [Contributing](.github/contributing) guide.

### How To Contribute

No contribution is too small, whether it be bug reports/fixes, feature requests, documentation updates, or anything else that can help drive the project forward.

Here are some good places to start:

* [Project Kanban Board](https://github.com/markphelps/flipt/projects/1)
* [Help Wanted](https://github.com/markphelps/flipt/labels/help%20wanted) Issues
* [Good First Issue](https://github.com/markphelps/flipt/labels/good%20first%20issue) Issues
* [Documentation](https://github.com/markphelps/flipt/labels/documentation) Issues

### Support Development

Or if you would like to support the continued development of Flipt (and my caffeine addiction), you could always [Buy Me A Coffee](https://www.buymeacoffee.com/mAZ1JDSRP)!

<a href="https://www.buymeacoffee.com/mAZ1JDSRP" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: auto !important;width: auto !important;" ></a>

## Pro Version

My plan is to soon start working on a Pro Version of Flipt for enterprise. Along with **support**, some of the planned features include:

* User management/permissions
* Multiple environments
* Audit log
* Streaming updates
* Metrics
* HA support

If you or your organization would like to help beta test a Pro version of Flipt, please get in touch with me:

* Twitter: [@mark_a_phelps](https://twitter.com/mark_a_phelps)
* Email: _mark.aaron.phelps at gmail.com_
