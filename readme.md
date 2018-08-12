tmt - Twitter Maintenance Tool
==============================

* `tmt followers`  ... list members you are followed
* `tmt followings`  ... list members you follows
* `tmt follow` ... follow person listed in STDIN
* `tmt dump IDNUM` ... dump JSON for the tweet
* `tmt post` ... tweet the contents of STDIN (utf8)

At first you runs tmt.exe , your default web-browser shows PIN number.
You have to write it into STDIN of tmt.exe.

How to build
------------

```
$ go get github.com/zetamatta/tmt
$ cd ~/go/src/github.com/zetamatta/tmt
$ cd secret
$ cp secret.go.sample secret.go
$ vim secret.go
```

```
package secret

const ConsumerKey = ""
const ConsumerSecret = ""
```

Write the values you get from https://apps.twitter.com/ , and

```
$ cd ..
$ go get ./...
$ go build
```

Author
------
[@zetamatta](https://github.com/zetamatta/)
