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

## Spellcheck

After making changes to the docs, it's a good idea to spellcheck them using the [markdown-spellcheck](https://www.npmjs.com/package/markdown-spellcheck) tool.

markdown-spellcheck can be installed with npm:

```shell
npm i markdown-spellcheck -g
```

To run a quick check, issue:

```shell
mdspell -r -n -a --en-us */**.md
```

This will output a list of potentially misspelled words. `mdspell` uses a `.spelling` file to track known words, so you may need to edit this file in the project root if `mdspell` reports a false positive.
