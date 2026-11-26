# `misc`

`misc` carries `car`, `coordinate`, `creditcard`
(Luhn-valid), `creditcard-test`, `cusip`, `currency` (ISO 4217), `datetime` (RFC 3339), `emoji`, `httpmethod`,
`httpstatus`, `iban-test`, `ip` (IPv4, or IPv6 one time in nine), `ipv6` (in `2001:db8::/32`), `isin`, `language` (ISO 639-1, with its 639-2/T code), `loglevel` (syslog
severity), `mac`, `mimetype`, `objectid`, `port`, `protocol`, `territory` (ISO
3166-1), `timezone` (IANA), `tld` (IANA root zone), `useragent`, `uuid` (v4) and `version` (such as `1.24.3` or `v2.0.1-rc.1`). Many
carry sub-fields — `misc.currency.symbol`, `misc.territory.alpha2`,
`misc.httpstatus.code` — which `--list` shows. `car`, `creditcard-test`, `currency`, `httpmethod`,
`httpstatus`, `iban-test`, `language`, `loglevel`, `mimetype`, `port`, `protocol`, `territory`,
`timezone`, `tld` and `useragent` are [tables](../../README.md#table), so
`misc.territory[SE].capital` and `misc.currency[Euro].symbol` select a row; `car` and
`useragent` carry no key or name, so they are drawn from rather than selected in.
`misc.useragent` carries `browser`, `device` and `os` beside the string, and `misc.car` a `make` and a `model`, until an international source replaces them.

## Reserved addresses

Five paths draw addresses set aside for documentation, private networks and local use. No vendor
holds one, and the public internet doesn't route one, though a private address can reach a host on
your own network:

- `misc.ip.documentation`: inside RFC 5737's `192.0.2.0/24`, `198.51.100.0/24` or `203.0.113.0/24`.
- `misc.ip.private`: inside RFC 1918's `10.0.0.0/8`, `172.16.0.0/12` or `192.168.0.0/16`.
- `misc.ip.cidr`: a block inside one of those six ranges, such as `192.168.7.0/24`.
- `misc.ipv6`: inside RFC 3849's `2001:db8::/32`, written as eight groups of one to four digits.
- `misc.mac.local`: a locally administered unicast address, whose first octet's second hex digit is `2`, `6`, `a` or `e`. A device may pick such an address for itself, so it can match a real one.

`documentation` and `private` never end in `.0` or `.255`. `misc.ip` and `misc.mac` draw addresses a real host may
hold; `misc.ip` takes them from `misc.ip.wide4`, an IPv4 address, or `misc.ip.wide6`, an IPv6 address.

## Credit cards and IBANs

A drawn `misc.creditcard` number or `{iban(CC)}` is valid, not real: it passes the Luhn check or
the IBAN's mod-97 check, and a real card or account may hold it. `{iban(CC)}` also carries the
check digits of the account number in BE, ES, FI and NO, and in SE the IBAN ID of a bank in
Bankinfrastruktur i Sverige's clearing number table. Every other digit is random, so a
checker that looks up a bank code, or the account digits behind an SE IBAN ID, or that checks
a DE or DK account, can still refuse a draw. `misc.creditcard-test` and
`misc.iban-test` draw only numbers published for testing: Stripe's test card for each brand
(`visa`, `mastercard`, `amex`, `discover`, `diners`, `jcb`, `unionpay`) and Stripe's test IBAN for
each of BE, DE, ES, FI, NO and SE, with Danske Bank's fictive example for DK. Each key holds one
number, so there are seven cards and seven IBANs, and `{/misc.creditcard-test[visa]}` always prints
`4242424242424242` and `{/misc.iban-test[SE]}` `SE3550000000054910000003`.

## Securities identifiers

`misc.isin` and `misc.cusip` are valid, not real: each passes the check its standard sets, and a
real security may hold the identifier. `misc.isin` is ISO 6166's twelve characters: an ISO 3166-1
alpha-2 code from `misc.territory`, nine random digits or capital letters, and the check digit
`{isin()}` draws. `misc.cusip` is ANSI X9.6's nine characters: eight random digits or capital
letters and the check digit `{cusip()}` draws. A register that looks the identifier up, or a
checker that requires a CUSIP's issuer to be registered, can still refuse a draw.

## Registry tables

`misc.httpmethod`, `misc.protocol` and `misc.port` are IANA's registries.
`misc.httpmethod` is the eight methods RFC 9110 defines and PATCH; the register's
other entries, WebDAV and DeltaV among them, do not ship. Its `safe` and
`idempotent` are `true` or `false`, so a Go `bool` reads them, though `--format json`
and `sql` write them as text, a table column carrying no datatype. `misc.protocol`
carries a `number` and a `name`, the keyword itself where the register spells none out,
and selects by either spelling.
`misc.port` is the TCP assignments, rendering the number a port field holds with the
IANA `service` beside it, and selecting by number alone. A draw spans the whole
register, so pin `misc.port[443]` where a fixture needs a port a reader recognises;
it stops at 49150, IANA assigning nothing above, so an ephemeral source port is
`{int(49152,65535)}`.
`misc.tld` is the root zone's delegated TLDs, keyed with the leading dot
`misc.territory.tld` already carries, `.se`. Its `type` is the register's —
`country-code`, `generic`, `generic-restricted`, `infrastructure` or `sponsored` — and
its `unicode` the form the register displays, `.рф` for `.xn--p1ai` and the key itself
for an ASCII one, which selects the row too, so `misc.tld[.рф]` renders `.xn--p1ai`. A
draw spans the whole zone, `.arpa` and `.zuerich` alike, so pin `misc.tld[.com]` where
a fixture needs one a reader recognises.
`misc.loglevel` is the eight syslog severities, rendering the keyword a configuration
writes, `err` and `info`, keyed by the numerical code a PRI encodes and selectable by
either, so `misc.loglevel[3]` and `misc.loglevel[err]` are one row.
`misc.loglevel.severity` is the name RFC 5424 spells, `Error` beside `err`; it
selects nothing, the code and the keyword do.

## Territories

The table is called `territory` because ISO 3166-1 codes territories, not sovereign
states: Greenland and Åland have codes of their own. `misc.territory.sovereign` names the
state a territory belongs to: `DK` for Greenland, `FI` for Åland. A sovereign state names
itself, and so does a territory its source names no state for, such as Western Sahara, `EH`.

## Time zones

`misc.timezone` is every zone tzdb gives a shipped territory, from one apiece for most
of them to dozens for the largest, weighted by the population GeoNames records in each,
so `misc.territory[US].timezone` draws `America/New_York` far more often than
`America/Nome`. It links to `misc.territory`, so `misc.territory[SE].timezone` is
`Europe/Stockholm` and a drawn territory and zone agree. Its `offset` is the zone's
*standard* offset, so do not pair it with a drawn `misc.datetime`: in a zone that
observes DST it is the wrong one half the year, which is what storing a zone name
avoids. A zone tzdb named recently — `Europe/Kyiv`, `America/Ciudad_Juarez` — is
rejected outright by a consumer resolving it against older tzdata, so where the
consumer validates the zone, pin one rather than draw it. There is no `UTC` row, tzdb
giving that name no territory — spell it as the text `"UTC"`.
