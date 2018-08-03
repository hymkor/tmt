tmt.exe - Twitter Maintenance Tool
==============================

```
$ tmt SUBCOMMANDS
```

SUBCOMMANDS
-----------

* tmt followers  ... list members you are followed
* tmt followings  ... list members you follows
* tmt follow ... follow person listed in STDIN
* tmt dump IDNum ... dump JSON for the tweet

At first you runs tmt.exe , your default web-browser shows PIN number.
You have to write it into STDIN of tmt.exe.

How to build
------------

```
$ cd secret
$ cp secret.go.sample secret.go
$ vim secret.go
```

```
package secret

const ConsumerKey = ""
const ConsumerSecret = ""
```

Write the values you get from https://apps.twitter.com/
