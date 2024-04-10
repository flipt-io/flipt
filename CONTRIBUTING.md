# Contributing

Checkout our [Development](DEVELOPMENT.md) guide for more information on how to get started developing Flipt.

## What To Work On

Check out our [public roadmap](https://github.com/orgs/flipt-io/projects/4) to see what we're working on and where you can help.

Not sure how to get started? You can:

- [Book a pairing session/code walkthrough](https://calendly.com/flipt-mark/30) with one of our teammates!
- Join our [Discord](https://www.flipt.io/discord), and ask any questions there

- Dive into any of the open issues, here are some examples: 
  - [Good First Issues](https://github.com/flipt-io/flipt/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)
  - [Backend](https://github.com/flipt-io/flipt/issues?q=is%3Aissue+is%3Aopen+label%3Ago)
  - [Frontend](https://github.com/flipt-io/flipt/issues?q=is%3Aopen+is%3Aissue+label%3Aui)

- Looking for issues by effort? We've got you covered:
  - [XS](https://github.com/flipt-io/flipt/issues?q=is%3Aissue+is%3Aopen+label%3Axs)
  - [Small](https://github.com/flipt-io/flipt/issues?q=is%3Aissue+is%3Aopen+label%3Asm)
  - [Medium](https://github.com/flipt-io/flipt/issues?q=is%3Aissue+is%3Aopen+label%3Amd)
  - [Large](https://github.com/flipt-io/flipt/issues?q=is%3Aissue+is%3Aopen+label%3Alg)
  - [XL](https://github.com/flipt-io/flipt/issues?q=is%3Aissue+is%3Aopen+label%3Axl)

## Issues

Let us know how we can help!

- Include any **stack traces** with your error
- List versions you are using: Flipt, Go, OS, etc.
- List the contents of your Flipt configuration file. (ex: default.yml)

## Code

It's always best to open a dialogue before investing a lot of time into a fix or new functionality.

Functionality must meet the design goals and vision for the project to be accepted; we would be happy to discuss how your idea can best fit into the future of Flipt.

Join our [Discord](https://www.flipt.io/discord) to chat with the team about any feature ideas or open a [Discussion](https://github.com/flipt-io/flipt/discussions) here on GitHub.

### Conventional Commits

We use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for commit messages. Please adhere to this specification when contributing.

## Testing

New functionality must have accompanying tests. We aim to keep the total test coverage of the project above 80%.

## Developer Certificate of Origin

We respect the intellectual property rights of others and we want to make sure
all incoming contributions are correctly attributed and licensed. A Developer
Certificate of Origin (DCO) is a lightweight mechanism to do that. The DCO is
a declaration attached to every commit. In the commit message of the contribution,
the developer simply adds a `Signed-off-by` statement and thereby agrees to the DCO,
which you can find below or at [DeveloperCertificate.org](http://developercertificate.org/).

```text
Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the
    best of my knowledge, is covered under an appropriate open
    source license and I have the right under that license to
    submit that work with modifications, whether created in whole
    or in part by me, under the same open source license (unless
    I am permitted to submit under a different license), as
    Indicated in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including
    all personal information I submit with it, including my
    sign-off) is maintained indefinitely and may be redistributed
    consistent with this project or the open source license(s)
    involved.
```

We require that every contribution to Flipt to be signed with a DCO. We require the
usage of known identity (such as a real or preferred name). We do not accept anonymous
contributors nor those utilizing pseudonyms. A DCO signed commit will contain a line like:

```text
Signed-off-by: Jane Smith <jane.smith@email.com>
```

You may type this line on your own when writing your commit messages. However, if your
user.name and user.email are set in your git configs, you can use `git commit` with `-s`
or `--signoff` to add the `Signed-off-by` line to the end of the commit message. We also
require revert commits to include a DCO.
