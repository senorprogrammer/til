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

## Usage

### Create a new page

```bash
❯ til Testing title
2020-04-20T14:52:57-07:00-testing-title.md
```

And then that page will open in [MacVim](https://macvim-dev.github.io/macvim/).

### Build the static pages

```bash
❯ til -build
```

And now the static pages are ready for committing up to GitHub. For example: [https://senorprogrammer.github.io/til/](https://senorprogrammer.github.io/til/).
