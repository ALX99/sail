# sail⛵

<p align="center">
  <img src="https://github.com/ALX99/sail/blob/master/sail.png" />
</p>

A non-bullshit, bug-free (as much as possible), minimal file manager for the terminal.

The goal of the project is to provide a simple, fast and easy to use file manager for the terminal. Under the hood it uses the great [bubbletea](github.com/charmbracelet/bubbletea) framework.

Sail strictly follows semantic versioning, meaning:

- patch version is incremented for bug fixes
- minor version is incremented for new features
- major version is incremented for breaking changes

so you can be confident that upgrading to a new version will not break your setup.

This project is inspired by the [lf](https://github.com/gokcehan/lf) file manager, which I previously used, but found to be too buggy.

## Preview

[![asciicast](https://asciinema.org/a/659987.svg)](https://asciinema.org/a/659987)

## Features

- [x] 24-bit (true color) support
- [x] LS_COLORS support
- [x] Customizable keybindings
- [x] *Sail* into directories
- [ ] Delete files
- [ ] Move files
- [ ] Copy files
- [ ] Rename files
- [ ] Create files
- [ ] Create directories
- [ ] Toggle hidden files
- [ ] Search files
- [ ] Open files with default application

## Usage

### Using sail as a cd replacement

Sail can be used as a replacement for the `cd` command. To do so, you can put the following in your `.bashrc` or `.zshrc`:

```sh
sd() {
  set -e
  tmp_file="$(mktemp)"
  trap 'rm -rf -- "$tmp_file"' EXIT INT TERM HUP
  sail -write-wd "$tmp_file"
  cd "$(cat "$tmp_file")"
  set +e
}
```

