<p align="center"><img src="till_header.png" width="916" height="306" alt="til" title="til: jot it down" /></p>

`til` is a fast, simple, command line-driven, mini-static site generator I use for quickly capturing and publishing one-off notes. Two commands, (three if you're picky about your commit messages).

Example output: [https://senorprogrammer.github.io/tilde/](https://senorprogrammer.github.io/tilde/)

# tl;dr

```bash
❯ til New title here
  ...edit
❯ til -save
```

And you're done.

# Contents

* [Installation](#installation)
	* [From source](#from-source)
* [Configuration](#configuration)
    * [Example](#config-example)
* [Usage](#usage)
    * [Creating a new page](#creating-a-new-page)
    * [Building static pages](#building-static-pages)
    * [Building, saving, committing, and pushing](#building-saving-committing-and-pushing)
* [Publishing to GitHub Pages](#publishing-to-github-pages)
* [Live Example](#live-example)

## Installation

### From source

```
go get -u github.com/senorprogrammer/til
cd $GOPATH/src/github.com/senorprogrammer/til
go install .
which til
til --help
```

## Configuration

When you first run `til --help` it will display the help and usage info. It also will also create a default configuration file. 

You will need to make some changes to this configuration file.

The config file lives in `~/.config/til/config.yml` (if you're an XDG kind of person, it will be wherever you've set that to).

Open `~/.config/til/config.yml`, change the following entries, and save it:

    * committerEmail
    * committerName
    * editor
    * targetDirectory
    
`committerEmail` and `committerName` are the values `til` will use to commit changes with when you run `til -save`. 

`editor` is the text editor `til` will open your file in when you run `til [some title here]`.

`targetDirectory` is where `til` will write your files to. If the target directory does not exist, `til` will try to create it. 

### Config Example

```
---
commitMessage: "build, save, push"
committerEmail: test@example.com
committerName: "TIL Autobot"
editor: "mvim"
targetDirectory: "~/Documents/til"
```

## Usage

`til` only has three usage options: `til`, `til -build`, and `til -save`.

### Creating a new page

```bash
❯ til New title here
2020-04-20T14-52-57-new-title-here.md
```

That new page will open in whichever editor you've defined in your config.

### Building static pages

```bash
❯ til -build
```

Builds the index and tag pages, and leaves them uncommitted.

### Building, saving, committing, and pushing

```bash
❯ til -save [optional commit message]
```

Builds the index and tag pages, commits everything to the git repo with the commit message you've defined in your config, and pushes it all up to the remote repo.

`-save` makes a hard assumption that your `targetDirectory` is under version control, controlled by `git`. It is highly recommended that you do this.

`-save` also makes a soft assumption that your `targetDirectory` has `remote` set to GitHub (but it should work with `remote` set to anywhere).

`-save` takes an optional commit message. If that message is supplied, it will be used as the commit message. If that message is not supplied, the `commitMessage` value in the config file will be used. If that value is not supplied, an error will be raised.

## Publishing to GitHub Pages

The generated output of `til` is such that if your `git remote` is configured to use GitHub, it should be fully compatible with GitHub Pages.

Follow the [GitHub Pages setup instructions](https://guides.github.com/features/pages/), using the `/docs` option for **Source**, and it should "just work".

## Live Example

An example published site: [https://senorprogrammer.github.io/tilde/](https://senorprogrammer.github.io/tilde/). And the raw source: [senorprogrammer/tilde](https://github.com/senorprogrammer/tilde)
