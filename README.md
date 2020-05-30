# Tangalanga

Zoom Conference scanner.

This scanner will check for a random meeting id and return information if available.

![](http://share.elcuervo.net/tangalanga-find-02.png)

## Usage

This are all the possible flags:

```bash
tangalanga \
    -token=user-token \   # [required]         user token to use

    -colors=false \       # [default: true]    enable/disable colors
    -output=history \     # [default: stdout]  write found meetings to file
    -debug=true \         # [default: false]   show all the attmpts
    -tor=true \           # [default: false]   enable tor connection (will use default socks proxy)
    -hidden=true \        # [default: false]   enable embedded tor connection (only linux)
    -rate=7 \             # [default: ncpu]    overwrite the default worker pool

    -proxy=socks5://... \ # [default: socks5://127.0.0.1:9150]   proxy url to use
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
