# mtimehash

[![Go Reference](https://pkg.go.dev/badge/github.com/slsyy/strintern.svg)](https://pkg.go.dev/github.com/slsyy/mtimehash)
[![Test](https://github.com/slsyy/mtimehash/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/slsyy/mtimehash/actions/workflows/test.yml)

CLI to modify files mtime (modification data time) based on the hash of the file content. This make it deterministic
regardless of when the file was created or modified.

## Instalation
```shell
go install github.com/slsyy/mtimehash/cmd/mtimehash@latest
```

## Rationale 

`go test` uses mtimes to determine, if files opened during tests has changed and thus: tests need to be re-run. 
Unfortunately in a typical CI workflow modifications times are random as `git` does not preserve them. This makes caching
for those tests ineffective, which slows down the test execution

The trick is to set mtime based on the file content hash. This way the mtime is deterministic regardless when the repository
was modified/clone, so hit ratio should be much higher.

## Usage

Pass a list of files to modify via stdin:

```shell
find . -type f | mtimehash
```