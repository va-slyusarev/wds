# WebDav Server (wds)

The utility provides access to the specified directory via the WebDAV protocol `without` authorization.

## Using

It is written to download video content to the iPad via the `AVPlayerHD` program.
It can be used by you for your own needs.

Install:
```shell script
go get github.com/va-slyusarev/wds
```

Use
```shell script
wds --help

  -d string
        WebDav server directory. (default ".")
  -p int
        WebDav server port. (default 80)
```