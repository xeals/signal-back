# signal-back

[![Build status](https://travis-ci.org/xeals/signal-back.svg?branch=master)](https://travis-ci.org/xeals/signal-back)

In version 4.17.5, the Signal Android app introduced encrypted backups. While these are undoubtedly a security benefit over unencrypted backups, they do present an issue in being read into other systems or simply by their owner.

`signal-back` is intended to use the same decryption process as the Signal app uses when importing its backups, to make them readable without being used by the app.

# Usage

```
Usage: signal-back COMMAND [OPTION...] BACKUPFILE

  --help, -h                show help
  --log FILE, -l FILE       write logging output to FILE
  --password PASS, -p PASS  use PASS as password for backup file
  --pwdfile FILE, -P FILE   read password from FILE
  --version, -v             print the version

Commands:
  format   Read and format the backup file
  analyse  Information about the backup file
  extract  Retrieve attachments from the backup
  help     Shows a list of commands or help for one command
```

The current interface is by no means complete or stable and may change in the future.

Current export formats are:
- XML: (hopefully) Compatible with [SMS Backup & Restore](https://play.google.com/store/apps/details?id=com.riteshsahu.SMSBackupRestore); however, there may be issues.
- CSV

Both formats currently only support SMS, though MMS is planned.

**Please create an issue if you have a format you'd like to see! (and the format is not already suggested)**

# Installing

Download whichever binary suits your system from the releases page; Windows, Mac OS (`darwin`), or Linux, and 32-bit (`386`) or 64-bit (`amd64`). Checksums are provided to verify file integrity.

## Building from source

Building requires [Go](https://golang.org) and [dep](https://github.com/golang/dep). If you don't have one (or both) of these tools, instructions should be easy to find. After you've initialised everything:

```
$ git clone https://github.com/xeals/signal-back $GOPATH/src/github.com/xeals/signal-back
$ cd $GOPATH/src/github.com/xeals/signal-back
$ dep ensure
$ go install .
```

You can also just use `go get github.com/xeals/signal-back`, but I provide no guarantees on dependency compatibility.

# Todo list

- [ ] Code cleanup
  - [ ] make code legible for other people
- [x] Actual command line-ness
- [x] Formatting ideas and options
- [ ] User-friendliness in errors and stuff

# License

Licensed under the Apache License, Version 2.0 ([LICENSE](LICENSE)
or http://www.apache.org/licenses/LICENSE-2.0).

## Contribution

Unless you explicitly state otherwise, any contribution intentionally submitted
for inclusion in the work by you, as defined in the Apache-2.0 license, shall be
licensed as above, without any additional terms or conditions.
