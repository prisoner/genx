# GenX : Generics For Go, Yet Again.
[![GoDoc](https://godoc.org/github.com/OneOfOne/genx?status.svg)](https://godoc.org/github.com/OneOfOne/genx)
[![Build Status](https://travis-ci.org/OneOfOne/genx.svg?branch=master)](https://travis-ci.org/OneOfOne/genx)
[![Report Card](https://goreportcard.com/badge/github.com/OneOfOne/genx)](https://goreportcard.com/report/github.com/OneOfOne/genx)

## Install

	go get github.com/OneOfOne/genx/...

## Features
* It can be *easily* used with `go generate`, from the command line or as a library.
* `cmd/genx` Uses local files, packages, and optionally uses `go get` (with the `-get` flag) if the remote package doesn't exist.
* You can rewrite, remove and change pretty much everything.
* Allows you to merge a package of multiple files into a single one.
* *Safely* remove functions and struct fields.
* Automatically passes all code through `x/tools/imports` (aka `goimports`).
* If you intend on generating files in the same package, you may add `// +build genx` to your template(s).
* Transparently handles [genny](https://github.com/cheekybits/genny)'s `generic.Type`.
* Supports a few [seeds](https://github.com/OneOfOne/genx/tree/master/seeds/).
* Adds build tags based on the types you pass, so you can target specific types (ex: `// +build genx_t_string` or `// +build genx_vt_builtin` )
* Automatically handles nil returns, will return the zero value of the type.
* Doesn't need modifying the source package if there's only one type involved.

## Examples:
### Package:

```
//go:generate genx -pkg ./internal/cmap -t KT=interface{} -t VT=interface{} -m -o ./cmap.go
//go:generate genx -pkg ./internal/cmap -n stringcmap -t KT=string -t VT=interface{} -fld HashFn -fn DefaultKeyHasher -s "cm.HashFn=hashers.Fnv32" -m -o ./stringcmap/cmap.go
```
* Input [cmap](https://github.com/OneOfOne):  [cmap.go](https://github.com/OneOfOne/cmap/blob/master/cmap.go) / [lmap.go](https://github.com/OneOfOne/cmap/blob/master/lmap.go)
* Merged output `map[interface{}]interface{}`: [cmap_iface_iface.go](https://github.com/OneOfOne/cmap/blob/master/cmap_iface_iface.go)
* Merged output `map[string]interface{}`: [stringcmap/cmap_string_iface.go](https://github.com/OneOfOne/cmap/blob/master/stringcmap/cmap_string_iface.go)

### Advanced type targeting:
**Input**:

* https://github.com/OneOfOne/cmap/blob/master/cmap_if_cmplx.go
* https://github.com/OneOfOne/cmap/blob/master/cmap_if_number_amd32.go
* https://github.com/OneOfOne/cmap/blob/master/cmap_if_number_amd64.go
* https://github.com/OneOfOne/cmap/blob/master/cmap_if_string.go
* https://github.com/OneOfOne/cmap/blob/master/cmap_if_other.go
* https://github.com/OneOfOne/cmap/blob/master/cmap.go
* https://github.com/OneOfOne/cmap/blob/master/lmap.go

**Output**:
* [`-t KT=interface{},VT=interface{}`](https://github.com/OneOfOne/cmap/blob/master/cmap_iface_iface.go)
* [`-t KT=string,VT=interface{}`](https://github.com/OneOfOne/cmap/blob/master/stringcmap/cmap_string_iface.go)
* [`-t KT=uint64,VT=interface{}`](https://github.com/OneOfOne/cmap/blob/master/u64cmap/cmap_u64_iface.go)
### Single File:
```bash
➤ genx -f github.com/OneOfOne/cmap/lmap.go -t "KT=string,VT=int" -fn "NewLMap,NewLMapSize=NewStringInt" -n main -v -o ./lmap_string_int.go
```

### Modifying an external library that doesn't specifically support generics:
Using [fatih](https://github.com/fatih)'s excellent [set](https://github.com/fatih/set) library:

```
# -fn IntSlice -fnStringSlice to remove unneeded functions.
➤ genx -pkg github.com/fatih/set -t 'interface{}=uint64' -fn IntSlice -fn StringSlice -v -n uint64set -o ./uint64set
```

### Target native types with a fallback: [seeds/sort](https://github.com/OneOfOne/genx/tree/master/seeds/sort)

```
➤ genx -seed sort -t T=string -n main
...
func SortStrings(s []string, reverse bool) { ... }
...
➤ genx -seed sort -t T=*pkg.OtherType -n main
...
func SortPkgTypes(s []string, less func(i, j int) bool) { ... }
...
```

### Sets: [seeds/set](https://github.com/OneOfOne/genx/tree/master/seeds/set)
```
package set

type KeyType interface{}

type KeyTypeSet map[KeyType]struct{}

func NewKeyTypeSet() KeyTypeSet { return KeyTypeSet{} }

func (s KeyTypeSet) Add(vals ...KeyType) {
	for i := range vals {
		s[vals[i]] = struct{}{}
	}
}

func (s KeyTypeSet) Has(val KeyType) (ok bool) {
	_, ok = s[val]
	return
}

func (s KeyTypeSet) Delete(vals ...KeyType) {
	for i := range vals {
		delete(s, vals[i])
	}
}

func (s KeyTypeSet) Keys() (out []KeyType) {
	out = make([]KeyType, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	return
}
```

* Command: `genx -seed set -t KeyType=string -fn Keys`

* Output:
```go
package set

type StringSet map[string]struct{}

func NewStringSet() StringSet { return StringSet{} }

func (s StringSet) Add(vals ...string) {
	for i := range vals {
		s[vals[i]] = struct{}{}
	}
}

func (s StringSet) Has(val string) (ok bool) {
	_, ok = s[val]
	return
}

func (s StringSet) Delete(vals ...string) {
	for i := range vals {
		delete(s, vals[i])
	}
}
```
## FAQ

### Why?
Mostly a learning experience, also I needed it and the other options available didn't do what I needed.

For Example I needed to remove a field from the struct and change all usage of it for [stringcmap](https://github.com/OneOfOne/cmap/tree/master/stringcmap).

## TODO
* Documentation.
* Documention for using the library rather than the commandline.
* Support package tests.
* Handle removing comments properly rather than using regexp.
* More seeds.
* ~~Add proper examples.~~
* ~~Support specialized functions by type.~~
* ~~Support removing structs and their methods.~~

## Credits
* Originally used the excellent [astrewrite](https://github.com/fatih/astrewrite) library by [Fatih](https://github.com/fatih).

## [BUGS](https://github.com/OneOfOne/genx/issues?utf8=%E2%9C%93&q=label%3Abug%20)

## Usage ([`cmd/genx`](https://github.com/OneOfOne/genx/tree/master/cmd/genx/main.go)):
```
➤ genx -h
NAME:
   genx - Generics For Go, Yet Again.

USAGE:
   genx [global options] command [command options] [arguments...]

VERSION:
   v0.5

AUTHOR:
   Ahmed <OneOfOne> W. <oneofone+genx <a.t> gmail <dot> com>

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --seed seed-name                  alias for -pkg github.com/OneOfOne/genx/seeds/seed-name
   --in file, -f file                file to process, use `-` to process stdin.
   --package package, --pkg package  package to process.
   --name name, -n name              package name to use for output, uses the input package's name by default.
   --type type, -t type              generic type names to remove or rename (ex: -t 'KV=string,KV=interface{}' -t RemoveThisType).
   --selector selector, -s selector  selectors to remove or rename (ex: -s 'cm.HashFn=hashers.Fnv32' -s 'x.Call=Something').
   --field field, --fld field        struct fields to remove or rename (ex: -fld HashFn -fld privateFunc=PublicFunc).
   --func func, --fn func            functions to remove or rename (ex: -fn NotNeededFunc -fn Something=SomethingElse).
   --out value, -o value             output dir if parsing a package or output filename if you want the output to be merged. (default: "/dev/stdout")
   --tags value                      go extra build tags, used for parsing and automatically passed to any go subcommands.
   --goFlags flags                   extra flags to pass to go subcommands flags (ex: --goFlags '-race')
   --get                             go get the package if it doesn't exist (default: false)
   --verbose, -v                     verbose output (default: false)
   --help, -h                        show help (default: false)
   --version, -V                     print the version (default: false)
```

## Contributions
* All contributions are welcome, just open a pull request.

## License

Apache v2.0 (see [LICENSE](https://github.com/OneOfOne/genx/blob/master/LICENSE) file).

Copyright 2016-2017 Ahmed <[OneOfOne](https://github.com/OneOfOne/)> W.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
