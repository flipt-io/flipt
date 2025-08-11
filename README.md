<img referrerpolicy="no-referrer-when-downgrade" src="https://static.scarf.sh/a.png?x-pxid=dc285028-90ba-436b-9a4a-e0d826a2c986" alt="" />

<p align=center>
    <img src="logo.svg" alt="Flipt" width=275 height=96 />
</p>

<p align="center">The Git-native, enterprise-ready feature management platform that developers love</p>

<hr />

<p align="center">
    <img src=".github/images/dashboard.png" alt="Flipt Dashboard" width=600 />
</p>

<br clear="both"/>

<div align="center">
    <!-- <a href="https://github.com/flipt-io/flipt/releases">
        <img src="https://img.shields.io/github/release/flipt-io/flipt.svg?style=flat" alt="Releases" />
    </a> -->
    <img src="https://img.shields.io/badge/status-stable-green" alt="Flipt v2 Stable" />
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
        <a href="https://docs.flipt.io/v2/introduction">Docs</a> •
        <a href="http://www.flipt.io">Website</a> •
        <a href="http://blog.flipt.io">Blog</a> •
        <a href="#contributing">Contributing</a> •
        <a href="https://www.flipt.io/discord">Discord</a>
    </h4>
</div>

> [!NOTE]  
> **Looking for Flipt v1?** You can find the v1 code on the [`main` branch](https://github.com/flipt-io/flipt/tree/main) and documentation at [docs.flipt.io](https://docs.flipt.io/).

## Why Flipt v2?

**Finally, feature flags that work with your existing Git workflow.**

Flipt v2 is the first truly Git-native feature management platform that treats your feature flags as code. Store your flags in your own Git repositories, use your existing branching strategy, and deploy flags alongside your code using the tools you already know and trust.

### 🚀 **Git-Native by Design**

- **Own your data**: Store feature flags directly in your Git repositories
- **Version control**: Full history and blame for every flag change
- **Branch and merge**: Test flag changes in branches before merging to production
- **Deploy together**: Feature flags deploy with your code using existing CI/CD pipelines

### 🌍 **Multi-Environment with Git Flexibility**

- **Environment per branch**: Map environments to Git branches for seamless workflows  
- **Environment per directory**: Organize flags by microservice or team within a single repo
- **Environment per repository**: Separate repos for different products or security domains
- **Complete isolation**: Each environment has its own namespaces, flags, and configurations

### ⚡ **Developer Experience First**

- **Zero infrastructure**: No databases, no external dependencies by default
- **GitOps ready**: Works with existing Git-based deployment workflows
- **Real-time updates**: Server-Sent Events (SSE) streaming API for instant flag propagation to client-side SDKs without polling
- **Modern UI**: Intuitive interface with full Git integration and dark mode support

### 🔒 **Enterprise Security & Control**

- **Self-hosted**: Keep sensitive flag data within your infrastructure
- **Secrets management**: Secure storage and retrieval of sensitive configuration data (OSS)
- **GPG commit signing**: Cryptographically sign all flag changes for enhanced security (Pro feature)
- **Merge proposals**: Code review workflow for flag changes (Pro feature)
- **Audit trails**: Complete history of who changed what and when
- **OIDC/JWT/OAuth**: Enterprise authentication methods supported

<br clear="both"/>

## Flipt v1 vs v2: What's New?

| Feature | Flipt v1 | ✨ Flipt v2 |
|---------|----------|----------|
| **Storage** | Database-centric (MySQL, PostgreSQL, SQLite) | Git-native with optional SCM sync (GitHub, GitLab, BitBucket, Azure DevOps, Gitea, etc.) |
| **Environments** | Single namespace model | Multi-environment with Git flexibility |
| **Branching** | Not supported | Full Git branching with environment branches |
| **Data Ownership** | Stored in a database (MySQL, PostgreSQL, SQLite) | Stored in your Git repositories alongside your code |
| **GitOps** | Read-only Git integration | Full read/write Git integration |
| **Deployment** | Requires database setup | Zero dependencies - single binary |
| **Version Control** | Basic audit logs | Full Git history and blame |
| **Merge Process** | Direct flag changes | Merge proposals with code review |
| **Real-time Updates** | Polling required | Server-Sent Events (SSE) streaming API for instant updates |
| **Multi-tenancy** | Manual namespace management | Environment-based isolation |
| **Secrets Management** | None | File-based providers available in OSS, HashiCorp Vault (Pro feature), cloud secrets management (AWS Secrets Manager, GCP Secret Manager, Azure Key Vault) coming soon |

<br clear="both"/>

## Use Cases

Perfect for engineering teams that want to:

- **Ship faster with confidence**: Deploy flags with your code using existing Git workflows
- **Enable trunk-based development**: Merge incomplete features behind flags without risk
- **Implement proper GitOps**: Treat infrastructure and feature flags as code
- **Maintain security compliance**: Keep sensitive flag data within your infrastructure  
- **Scale with multiple environments**: Manage flags across dev, staging, and production seamlessly
- **Enable team collaboration**: Use familiar Git workflows for flag reviews and approvals

<br clear="both"/>

## Quick Start

See our [quickstart guide](https://docs.flipt.io/v2/quickstart) for more details.

### Local

```bash
# Install Flipt
curl -fsSL https://get.flipt.io/v2 | sh

# Wizard-driven setup to get you started quickly
flipt quickstart

# Run Flipt server
flipt server
```

### Docker

```bash
docker run --rm -p 8080:8080 -p 9000:9000 -t docker.flipt.io/flipt/flipt:v2
```

Flipt UI will be available at [http://127.0.0.1:8080/](http://127.0.0.1:8080).

### Configuration Example

```yaml
# config.yml - Git-native setup with secrets management
secrets:
  providers:
    # File-based secrets (OSS)
    file:
      enabled: true
      base_path: "/etc/flipt/secrets"
    # HashiCorp Vault (Pro feature)
    vault:
      enabled: true
      address: "https://vault.example.com"
      auth_method: "token"
      token: "hvs.your_token"
      mount: "secret"

storage:
  type: git
  git:
    repository: "https://github.com/your-org/feature-flags.git"
    ref: "main"
    poll_interval: "30s"
    signature:
      enabled: true
      type: "gpg"
      key_ref:
        provider: "vault"  # Requires Pro license
        path: "flipt/signing-key"
        key: "private_key"
      name: "Flipt Bot"
      email: "bot@example.com"

environments:
  default:
    storage: git
  staging:
    storage: git
    directory: "staging"
```

For more setup options, see our [configuration documentation](https://docs.flipt.io/v2/configuration/overview).

<br clear="both"/>

## Core Values

- 🔒 **Security** - HTTPS, OIDC, JWT, OAuth, K8s Service Token, and API Token authentication methods supported out of the box
- 🗝️ **Secrets Management** - Secure storage and retrieval of sensitive data with file-based providers (OSS) and HashiCorp Vault (Pro feature), with upcoming cloud provider support (AWS, GCP, Azure)
- 🎛️ **Control** - Your data stays in your Git repositories within your infrastructure  
- 🚀 **Speed** - Co-located with your services, no external API calls required
- ✅ **Simplicity** - Single binary with no external dependencies by default
- 🔄 **GitOps Ready** - Native Git integration that works with your existing workflows
- 👍 **Compatibility** - GRPC, REST, Redis, Prometheus, ClickHouse, OpenTelemetry, and more

<br clear="both"/>

## Key Features

### Git-Native Storage

- Store flags directly in Git repositories alongside your code
- Full version control with Git history, blame, and diff support  
- Integrates with your SCM (GitHub, GitLab, BitBucket, Azure DevOps, Gitea, etc.)
- GPG commit signing for cryptographic verification of changes

### Multi-Environment Management  

- Environment per Git branch, directory, or repository
- Complete environment isolation with independent configurations
- Seamless environment promotion workflows

### Secrets Management & Security

- **Multi-provider secrets management**: File-based providers available in OSS, HashiCorp Vault (Pro feature), with AWS Secrets Manager, GCP Secret Manager, and Azure Key Vault support coming soon
- **GPG Commit Signing**: Cryptographically sign all flag changes with keys from secret providers (Pro feature)
- **Secure key storage**: Private keys and sensitive data stored securely in Vault or local files
- **Multiple auth methods**: Token, Kubernetes, and AppRole authentication for Vault

### Advanced Flag Management

- Complex targeting rules and user segmentation
- Percentage-based rollouts
- Real-time flag evaluation with Server-Sent Events (SSE) streaming updates for instant synchronization

### Developer Experience

- Modern UI with Git integration and dark mode 🌙
- Declarative flag configuration with JSON/YAML schemas
- Comprehensive REST and gRPC APIs

### Enterprise & Security Features

- **Authentication**: OIDC, JWT, OAuth, and more authentication methods supported out of the box
- **Observability**: OpenTelemetry and Prometheus integration 🔋
- **Secrets Management**: File-based providers (OSS) with HashiCorp Vault integration available in Pro
- **Enterprise Workflows**: GPG commit signing, merge proposals, and DevOps integration available in Pro

> **Ready for enterprise features?** Flipt v2 Pro includes advanced workflows, enhanced security, and dedicated support. Get started with a **free 14-day trial** – no credit card required.
>
> **[Learn More About Pro Features →](https://docs.flipt.io/v2/pro)**

Are we missing a feature that you'd like to see? [Let us know by opening an issue!](https://github.com/flipt-io/flipt/issues)

<br clear="both"/>

## Integration & SDKs

Check out our [integration documentation](https://docs.flipt.io/v2/integration/overview) for comprehensive guides.

### Server-Side Evaluation

- **REST API** - Full HTTP API for any language
- **gRPC API** - High-performance binary protocol

### Client-Side Evaluation  

- **Local evaluation** - Reduce latency with client-side flag evaluation, evaluate flags within your application for extreme speed and reliability.

### OpenFeature Integration

Flipt supports the [OpenFeature](https://openfeature.dev/) standard for vendor-neutral feature flag evaluation, including the emerging OpenFeature Remote Evaluation Protocol (OFREP) for standardized remote flag evaluation.

<br clear="both"/>

## Contributing

We would love your help! Before submitting a PR, please read over the [Contributing](CONTRIBUTING.md) guide.

No contribution is too small, whether it be bug reports/fixes, feature requests, documentation updates, or anything else that can help drive the project forward.

Not sure how to get started? You can:

- Join our [Discord](https://www.flipt.io/discord), and ask any questions there
- Dive into any of the open issues, here are some examples:
  - [Good First Issues](https://github.com/flipt-io/flipt/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)
  - [Backend](https://github.com/flipt-io/flipt/issues?q=is%3Aissue+is%3Aopen+label%3Ago)
  - [Frontend](https://github.com/flipt-io/flipt/issues?q=is%3Aopen+is%3Aissue+label%3Aui)

Review the [Development](DEVELOPMENT.md) documentation for more information on how to contribute to Flipt.

<br clear="both"/>

## Flipt v2 Pro

Ready to unlock the full potential of Git-native feature management? Flipt v2 Pro adds enterprise-grade features on top of our solid open-source foundation.

### What's Included in Pro

- **🔀 Enterprise DevOps Integration** - Native GitHub, GitLab, BitBucket, Azure DevOps, and Gitea workflows with merge proposals and automated PR/MR creation
- **✍️ GPG Commit Signing** - Cryptographically sign all changes for maximum security and auditability with verified commit badges
- **🔒 Integrated Secrets Management** - Built-in secure storage for GPG keys with HashiCorp Vault integration and upcoming cloud secrets management support (AWS Secrets Manager, GCP Secret Manager, Azure Key Vault)
- **🌐 Air-Gapped Environment Support** - Annual licenses work seamlessly in air-gapped environments with offline validation
- **🏢 Dedicated Support Channel** - Direct access to our engineering team via dedicated Slack channel for faster issue resolution
- **⚡ Priority Development** - Your bug reports and feature requests get prioritized
- **🔧 Enterprise Auth** - Advanced authentication providers (🔄 coming soon)
- **📊 Advanced Analytics** - Enhanced reporting and insights (🔄 coming soon)

### Pricing & Trial

- **Free 14-day trial** - No credit card required to start, includes all Pro features
- **Monthly licenses** - $100/month with continuous internet connectivity for validation
- **Annual licenses** - $1,000/year (save $200) with offline validation using license files for air-gapped environments
- **No instance limits** on paid plans - run Flipt v2 Pro on as many servers as you need
- **Cancel anytime** - Prorated billing through our Stripe customer portal

**[Start Your Free 14-Day Trial →](https://docs.flipt.io/v2/pro#getting-started)**

*Trial includes up to 5 instances. Upgrade seamlessly to unlimited instances with a paid subscription.*

<br clear="both"/>

## Community

For help and discussion around Flipt, feature flag best practices, and more, join us on [Discord](https://www.flipt.io/discord).

<br clear="both"/>

<p align="center">
    <a href="https://www.producthunt.com/posts/flipt-cloud?embed=true&utm_source=badge-featured&utm_medium=badge&utm_souce=badge-flipt&#0045;cloud" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=498373&theme=light" alt="Flipt&#0032;Cloud - Feature&#0032;flags&#0046;&#0032;Powered&#0032;by&#0032;Git | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>
    <a href="https://devhunt.org/tool/flipt" title="DevHunt - Tool of the Week"><img src="./.github/images/devhunt-badge.png" width=225 alt="DevHunt - Tool of the Week" /></a>&nbsp;
    <a href="https://console.dev/tools/flipt" title="Visit Console - the best tools for developers"><img src="./.github/images/console-badge.png" width=250 alt="Console - Developer Tool of the Week" /></a>
</p>

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

## Licensing

There are currently two types of licenses in place for Flipt:

1. Client License  
2. Server License

### Client License

All of the code required to generate GRPC clients in other languages as well as the [Go SDK](/sdk/go) are licensed under the [MIT License](https://spdx.org/licenses/MIT.html).

This code exists in the [rpc/](rpc/) directory.

The client code is the code that you would integrate into your applications, which is why a more permissive license is used.

### Server License

The server code is licensed under the [Fair Core License, Version 1.0, MIT Future License](https://github.com/flipt-io/flipt/blob/v2/LICENSE).

See our [licensing docs](https://docs.flipt.io/v2/licensing) and [fcl.dev](https://fcl.dev) for more information.
