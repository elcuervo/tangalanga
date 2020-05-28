# Tangalanga

Zoom Conference scanner.

This scanner will check for a random meeting id and return information if available.

![](http://share.elcuervo.net/tangalanga-find-02.png)

## Usage

This are all the possible flags:

```bash
tangalanga \
    -token=user-token \ # User token to user
    -output=log/history \ # Write founds on file
    -debug=true \ # Show not founds too
    -tor=true # Enable tor connection (not available on windows and osx)
```

## Tokens

Unfortunately I couldn't find the way the tokens are being generated but the core concept is that
the `zpk` cookie key is being sent during a Join will be usable for ~24 hours before expiring. This
makes trivial to join several known meetings, gether some tokens and then use them for the scans.

## TOR (only linux)

Tangalanga has a tor runtime embedded so it can connect to the onion network and run the queries
there instead of exposing your own ip.

![](http://share.elcuervo.net/tangalanga-find-tor-01.png)

For any other system I recommend a VPN

## Why the bizarre name?

This makes reference to a famous 80s/90s personality in the Rio de la Plata. [Doctor Tangalanga](https://en.wikipedia.org/wiki/Dr._Tangalanga)
who loved to do phone pranks.
