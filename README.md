# Silrok Feeder for SoftLayer

This is silrok feeder for softlayer tickets.


## Feature

Silrok feeder for softlayer does following things:

* Grab all new tickets via SoftLayer API.
* Archive all tickets into Elastic Stack via Logstash.
* Send alert messages to Slack channels for interesting accounts.
  * It does filtering with followings:
    * Just send alerts for specified accounts with command line arguments.
    * Do not send alerts for user-generated tickets.
    * Do not send alerts for well-known safe messages. (hard-coded)

### Well-known Safe Messages

Following messages are ignored when making an alert:

* Automated sales message like "Server Cancellation ..."
* Automated statistics message like "[ABNS] ..." (network usage)
* Automated configuation updates like "Customer Upgrade Request..."


## Installation

It requires basic build tools for Go application. If you already have one,
installation of silrok feeder for softlayer is quite simple!

### Get the Source and Build!

With `go get` command, you can get the source tree of the feeder and build it
at once.

```console
$ go get -u github.com/hardenedlayer/silrok-feeder-softlayer/cmd/silrok-feeder-softlayer
$ ls $GOPATH/bin/silrok-feeder-softlayer
<...>/bin/silrok-feeder-softlayer
$
```

Yes, you have one.


## Run!

Silrok feeder for softlayer is stand-alone command line program.

### Options

Before start, check supported options.

```console
$ ./silrok-feeder-softlayer  -o
Usage: silrok-feeder-softlayer [-os] [-a value] [-f value] [-h value] [-k value] [-u value] [parameters ...]
 -a, --addr=value  Address of Archiver
 -f, --from=value  Start date of fetching (YYYY-MM-DD)
 -h, --hook=value  Webhook URL for Slack
 -k, --apikey=value
                   API Key of SoftLayer Brand Account
 -o, --options     Show options
 -s, --fetchall    Sync All Tickets
 -u, --user=value  API Username of SoftLayer Brand Account
$ 
```

All options can be omitted but then it will not works! What? :-)

Options `-u` and `-k` are used to specify authentication information for
SoftLayer API. They cannot be omitted. Since this feeder needs brand
permission to get all tickets of all accounts under the brand, the user
and API key should be a set of brand agent's.

Option `-h` specify URL of slack webhook for slack alert feature. If this
option is omitted, slack alert feature will be disabled automatically.
*Note that you can only use single webhook for the process. It does not
support multiple webhook URLs for each channels.*

If you also want to enable archiving mode, option `-a` is required and it
should point the network address of archiving server. If this option is
omitted, the feature will be disabled automatically.

When you use the feeder for retrieving all tickets from specific date,
especially for archiving, you can use option `-s` and `-f`. Option `-f`
specify the starting date for archiving. Note that, without `-s`, the
option `-f` will not work. (Why? Hmm...)


OK, let's go!

### Run as Archiver

Following command will fetch and archive all tickets from Jan 11st 2018 and
continue until you break its running.

```console
$ silrok-feeder-softlayer -u agent -k 3jdf94jfakd<...> -a <ip_of_logstash> -s -f 2018-01-11
<...>
^C
$ 
```


### Run as Alert Feeder

Following command will fetch all new tickets from its starting time and
make alerts for account number 0000 and 0001, then send them to channel
`#stark` for 0000 and `#widow` for 0001.

```console
$ silrok-feeder-softlayer -u agent -k ek3jfj3kdf9<...> -h <hook_url> 0000:stark 0001:widow
<...>
^C
$ 
```


### All Together Now

Yes, you don't need separated process for both archiving and alerting.

```console
$ silrok-feeder-softlayer -u agent -k 3jdf94jfakd<...> -a <ip_of_logstash> -h <hook_url> 0000:stark 0001:widow
<...>
^C
$ 
```

Note that, do not use `-s` and `-f` for this case. If you use them when the
alert feature is enabled by `-h`, Oops! It kills all channel members!


## TODO

* Integrate it into Alargo and make it as period job.


## Author

Yonghwan SO https://github.com/sio4

## Copyright (GNU General Public License v3.0)

Copyright 2016 Yonghwan SO

This program is free software; you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation; either version 3 of the License, or (at your option) any later
version.

This program is distributed in the hope that it will be useful, but WITHOUT
ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with
this program; if not, write to the Free Software Foundation, Inc., 51
Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA

