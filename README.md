# til

A place to keep track of things I've learned that don't warrant a blog post.

I use this `zsh` alias to execute it from whichever directory I'm in:

```shell
alias til='cd ~/Documents/til && go run ./til.go'
```

## Usage

### Creating a new page

```bash
❯ til Testing title
2020-04-20T14:52:57-07:00-testing-title.md
```

### Building static pages

```bash
❯ til -build