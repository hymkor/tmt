tmt - Twitter Maintenance Tool
==============================

* show timeline
    * `tmt timeline` ... get home timeline
    * `tmt mention` ... get mention timeline
    * `tmt said` ... show what I said
    * `tmt dump IDNum` ... dump JSON for the tweet
* post
    * `tmt post` ... tweet the contents of STDIN (utf8)
    * `tmt cont` ... tweet in the same thread as the last tweeting
    * `tmt reply IDNum` ... reply to IDNum
* maintainance
    * `tmt followers`  ... list members you are followed
    * `tmt followings`  ... list members you follows
    * `tmt follow` ... follow person listed in STDIN
    * `tmt unfollow` ... follow person listed in STDIN
    * `tmt whoami` ... show who are you

At first you runs `tmt` , your default web-browser shows PIN number.
You have to write it into STDIN of tmt.exe.

In STDIN, please write username like `@ScreenName` per one line.

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
