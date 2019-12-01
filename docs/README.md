# Flipt Docs

The docs directory contains the markdown files used to build the Flipt documentation.

The documentation is built using [MkDocs](https://www.mkdocs.org) using the [mkdocs-material](https://github.com/squidfunk/mkdocs-material) theme.

After making changes to any of the markdown files in the docs directory, run `make docs` to rebuild the documentation.

## Build Requirements

* Python 2.7
* pip

## Build Setup

```bash
pip install mkdocs mkdocs-material pygments
```

## MKDocs

Once you have installed the necessary dependencies, you can run (from project root):

```shell
mkdocs serve
```

This will generate and serve the documentation locally at: `http://localhost:8000`.
