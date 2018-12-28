<p align="center">ircdef, IRC constants and definitions, in Go!</p>
<p align="center">
  <a href="https://travis-ci.org/lrstanley/ircdef"><img src="https://travis-ci.org/lrstanley/ircdef.svg?branch=master" alt="Build Status"></a>
  <a href="https://godoc.org/github.com/lrstanley/ircdef"><img src="https://godoc.org/github.com/lrstanley/ircdef?status.png" alt="GoDoc"></a>
  <a href="https://goreportcard.com/report/github.com/lrstanley/ircdef"><img src="https://goreportcard.com/badge/github.com/lrstanley/ircdef" alt="Go Report Card"></a>
  <a href="https://byteirc.org/channel/%23%2Fdev%2Fnull"><img src="https://img.shields.io/badge/ByteIRC-%23%2Fdev%2Fnull-blue.svg" alt="IRC Chat"></a>
</p>

## Table of Contents
- [Summary](#summary)
- [Status](#status)
- [Packages](#packages)
- [TODO](#todo)
- [Contributing](#contributing)
- [License](#license)

## Summary

This package contains a set of IRC definition lists. Things like IRC numerics,
modes (channel, user, etc), and other adaptions that IRC server software have
implemented. This package is generated using the code in the `codegen/` folder.

The data used to generate this package is obtained from
[ircdocs/ircdefs](https://github.com/ircdocs/irc-defs), and
[alien.net.au](https://www.alien.net.au/irc/irc2numerics.html). Please check
out [defs.ircdocs.horse](https://defs.ircdocs.horse/) for html-generated
documentation on these numerics/modes/etc.

## Status

These packages are still a work in progress and may change, please do not use
them yet. In the future, this will be automatically updated every day if there
are any changes to the source dataset.

## Packages

Only some of the dataset was used so far. For each dataset, take a look at
the following table:

| dataset | folder | source | code docs | html docs |
| ------- | ------ | ------ | --------- | --------- |
{{- range $k,$v := .data }}
| **{{ $v.Data.Page.Name }}** | [`{{ $k }}/`]({{ $k }}/) | [{{ $v.Path }}](https://github.com/ircdocs/irc-defs/blob/gh-pages/_data/{{ $k }}.yaml) | [docs](https://godoc.org/github.com/lrstanley/ircdef/{{ $k }}) | [{{ $k }}.html](https://defs.ircdocs.horse/defs/{{ $k }}.html) |
{{- end }}

## TODO

- [ ] camelcase vs original CAP_CASE? (tough because some things can't automatically be converted, e.g. `NOKNOCK` becomes `Noknock` but `NO_KNOCK` becomes `NoKnock`..)
- [ ] https://github.com/iancoleman/strcase/issues/13 (if this is added, could use it.)

## Contributing

Please review the [CONTRIBUTING](CONTRIBUTING.md) doc for submitting issues/a
guide on submitting pull requests and helping out.

## License

Note: some of the source code in this repository is copyright the respective
owners of the datasets used to generate those files. For these scenarios,
please refer to the header of the file for copyright information.

```
MIT License

Copyright (c) 2018 Liam Stanley <me@liamstanley.io>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
