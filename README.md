<p align="center"><img src="images/till_header.png" width="916" height="306" alt="til" title="til: jot it down" /></p>

`til` is a fast, simple, command line-driven, mini-static site generator for quickly capturing and publishing one-off notes. 

All in only two commands.

Example output: [https://github.com/senorprogrammer/tilde](https://github.com/senorprogrammer/tilde)

[![Go Report Card](https://goreportcard.com/badge/github.com/senorprogrammer/til)](https://goreportcard.com/report/github.com/senorprogrammer/til)

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
	* [As a binary](#as-a-binary)
* [Configuration](#configuration)
    * [Example](#config-example)
* [Usage](#usage)
    * [Creating a new page](#creating-a-new-page)
    * [Building static pages](#building-static-pages)
    * [Building, saving, committing, and pushing](#building-saving-committing-and-pushing)
* [Publishing to GitHub Pages](#publishing-to-github-pages)
* [Live Example](#live-example)
* [Frequently Unasked Questions](#frequently-unasked-questions)

## Installation

### From source

```
go get -u github.com/senorprogrammer/til
cd $GOPATH/src/github.com/senorprogrammer/til
go install .
which til
til --help
```

### As a Binary

[Download the latest binary](https://github.com/senorprogrammer/til/releases) from GitHub.

til is a stand-alone binary. Once downloaded, copy it to a location you can run executables from (ie: `/usr/local/bin/`), and set the permissions accordingly:

```bash
chmod a+x /usr/local/bin/til
```

and you should be good to go.

## Configuration

When you first run `til --help` it will display the help and usage info. It also will also create a default configuration file. 

You will need to make some changes to this configuration file.

The config file lives in `~/.config/til/config.yml` (if you're an XDG kind of person, it will be wherever you've set that to).

Open `~/.config/til/config.yml`, change the following entries, and save it:

    * committerEmail
    * committerName
    * editor
    * targetDirectories
    
`committerEmail` and `committerName` are the values `til` will use to commit changes with when you run `til -save`. 

`editor` is the text editor `til` will open your file in when you run `til [some title here]`.

`targetDirectories` defines the locations that `til` will write your files to. If a specified target directory does not exist, `til` will try to create it. This is a map of key/value pairs, where the "key" defines the value to pass in using the `-target` flag, and the "value" is the path to the directory.

If only one target directory is defined in the configuration, the `-target` flag can be ommitted from all commands. 
If multiple target diretories are defined in the configuration, all commands must include the `-target` flag specifying 
which target directory to operate against.

### Config Example

```
---
commitMessage: "build, save, push"
committerEmail: test@example.com
committerName: "TIL Autobot"
editor: "mvim"
targetDirectories: 
    a: ~/Documents/notes
    b: ~/Documents/blog
```

## Usage

`til` only has three usage options: `til`, `til -build`, and `til -save`.

### Creating a new page

With one target directory defined in the configuration:

```bash
❯ til New title here
2020-04-20T14-52-57-new-title-here.md
```

With multiple target directories defined:

```bash
❯ til -target a New title here
2020-04-20T14-52-57-new-title-here.md
```

That new page will open in whichever editor you've defined in your config.

### Building static pages

With one target directory defined in the configuration:

```bash
❯ til -build
```

With multiple target directories defined:

```bash
❯ til -target a -build
```

Builds the index and tag pages, and leaves them uncommitted.

<p align="center"><img src="images/til_build.png" width="600" height="213" alt="image of the build process" title="til -build" /></p>

### Building, saving, committing, and pushing

With one target directory defined in the configuration:

```bash
❯ til -save [optional commit message]
```

With multiple target directories defined:

```bash
❯ til -target a -save [optional commit message]
```

Builds the index and tag pages, commits everything to the git repo with the commit message you've defined in your config, and pushes it all up to the remote repo.

`-save` makes a hard assumption that your target directory is under version control, controlled by `git`. It is recommended that you do this.

`-save` also makes a soft assumption that your target directory has `remote` set to GitHub (but it should work with `remote` set to anywhere).

`-save` takes an optional commit message. If that message is supplied, it will be used as the commit message. If that message is not supplied, the `commitMessage` value in the config file will be used. If that value is not supplied, an error will be raised.

<p align="center"><img src="images/til_save.png" width="600" height="259" alt="image of the save process" title="til -save" /></p>

## Publishing to GitHub Pages

The generated output of `til` is such that if your `git remote` is configured to use GitHub, it should be fully compatible with GitHub Pages.

Follow the [GitHub Pages setup instructions](https://guides.github.com/features/pages/), using the `/docs` option for **Source**, and it should "just work".

## Live Example

An example published site: [https://senorprogrammer.github.io/tilde/](https://senorprogrammer.github.io/tilde/). And the raw source: [github.com/senorprogrammer/tilde](https://github.com/senorprogrammer/tilde)

## Frequently Unasked Questions

### Isn't this just (insert your favourite not this thing here)?

Yep, probably. I'm sure you could also put something like this together with [Hugo](https://gohugo.io), or [Jekyll](https://jekyllrb.com), or bash scripts, or emacs and some lisp macros.... Cool, eh?

### Does it have search?

It does not.

### Does this work on Windows?

Good question. No idea. Let me know?
