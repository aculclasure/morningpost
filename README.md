# The Morning Post

`morningpost` is a CLI tool that curates a "morning newspaper" for you! By default, it gets stories from the [HackerNews API](https://github.com/HackerNews/API) but can easily be extended to bring in news from other sources as well.

## Using `morningpost`
  - Install `morningpost`
  - Get the latest 10 news stories:

```bash
$ morningpost

Latest HackerNews Stories
=========================

Aliens of New York, AI Art – Mourning
https://aony.substack.com/p/mourning

The Quicksort of Science
https://www.exfatloss.com/p/potato-riff-study-the-quicksort-of

BCHS software stack: BSD, C, httpd, SQLite
https://learnbchs.org/

Four ways to bring 'open' to business
https://opensource.com/open-organization/17/11/4-ways-open-stakeholders
  ...
```

## Installation

### From source

- Clone the repo.
- Install the `morningpost` binary.

```bash
cd ./cmd/morningpost && go install .
```

## Description

Full project description and instructions [link](./INSTRUCTIONS.md).