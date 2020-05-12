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

To use this yourself, the simplest approach is probably to fork this repo yourself and just delete the content from the `/docs` directory.

You could also:

* create your own empty git repo
* copy `til.go` into it
* create the `docs` directory: `mkdir docs`
* push that up to GitHub

Now run `til --help` to initialize everything and make sure it's working.

## Configuration

When you first ran `til --help` it will have either exploded with an error message (open an issue here with the message, if you like), or it will have displayed the help info. If you saw help info, it also will have created a configuration file that you'll need to edit.

That file is probably in `~/.config/til/config.yml`. If you're an XDG base directory kind of person, it will be wherever that is.

Open `~/.config/til/config.yml`, change the following, and save:

    * committerEmail
    * committerName
    * editor
    
`committerEmail` and `committerName` are the values it will use to commit with when running `til -build`. 

`editor` is the text editor it will open for writing in when running `til [some title here]`.

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

For example: [https://senorprogrammer.github.io/til/](https://senorprogrammer.github.io/til/).
