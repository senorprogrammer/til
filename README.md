<p align="center">
    <img src="./till.jpg?raw=true" title="till" alt="WTF" width="400" height="200" />
</p>

# til

A place to keep track of things I've learned that don't warrant a blog post.

I use this `zsh` alias to execute it from whichever directory I'm in:

```shell
alias til='cd ~/Documents/til && go run ./til.go'
```

(Yep, I don't bother to compile/install it, it's fast enough as-is).

## Installation

To use this yourself, the simplest approach is probably to clone this repo yourself and just delete the content from the `/docs` directory.

You could also:

* create your own empty git repo
* copy `til.go` into it
* create the `docs` directory: `mkdir docs`
* push that up to GitHub

**Note:** Don't forget to change the `commiterName` and `commiterEmail` constants in `til.go` to your own values. And if you don't use MacVim, change `editor` as well.

## Usage

### Creating a new page

```bash
❯ til Testing title
2020-04-20T14-52-57-testing-title.md
```

And then that page will open in [MacVim](https://macvim-dev.github.io/macvim/).

### Building static pages

```bash
❯ til -build
```

Builds the index and tag pages.

### Saving all

```bash
❯ til -save
```

Builds the index and tag pages, commits everything to the git repo with a generic commit message, and pushes it all up to the remote repo.

And now the static pages are ready for committing up to GitHub. For example: [https://senorprogrammer.github.io/til/](https://senorprogrammer.github.io/til/).
