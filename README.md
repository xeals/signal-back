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

The current interface is by no means complete and I intend to expand on it.

Currently only an XML output format is (partially) available. This attempts to be compatible with the [SMS Backup & Restore](https://play.google.com/store/apps/details?id=com.riteshsahu.SMSBackupRestore) app that is commonly used to back up the system SMS database. However, it is yet untested.

**Please contribute on [#2](https://github.com/xeals/signal-back/issues/2) if you have a format you'd like to see!**

# Installing

Currently installation is only manual because it's not ready for a release. I am to support Windows, MacOS, and Linux eventually.

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
- [ ] Formatting ideas and options
- [ ] User-friendliness in errors and stuff

# License

Licensed under the Apache License, Version 2.0 ([LICENSE](LICENSE)
or http://www.apache.org/licenses/LICENSE-2.0).

## Contribution

Unless you explicitly state otherwise, any contribution intentionally submitted
for inclusion in the work by you, as defined in the Apache-2.0 license, shall be
licensed as above, without any additional terms or conditions.
