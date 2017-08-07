# vsftpdmgr [![CircleCI](https://circleci.com/gh/amenzhinsky/vsftpdmgr.svg?style=svg)](https://circleci.com/gh/amenzhinsky/vsftpdmgr)

[vsftpd](https://en.wikipedia.org/wiki/Vsftpd) users management daemon.

## API

Create new or update existing user:
```
$ curl localhost:8181/users/ -d '{
  "username": "test",
  "password": "test",
  "fs": {
    "mode": 365,
    "owner": "ftp",
    "group": "ftp",
    "children": [
      {
        "name": "read",
        "mode": 365,
        "owner": "ftp",
        "group": "ftp"
      },
      {
        "name": "write",
        "mode": 493,
        "owner": "ftp",
        "group": "ftp"
      }
    ]
  }
}'
```

JSON doesn't support octals, so 0555 is 365 and 0755 is 493.

Delete user:
```
$ curl -X DELETE localhost:8181/users/ -d '{"username": "test"}'
```

List all users:
```
$ curl localhost:8181/users/
```

## Running

The service requires a database storage, but currently only postgresql is supported.

`DATABASE_URL` is passed to the binary via environment variable instead of a flag for security reasons.

```
$ export DATABASE_URL=postgres://user:pass@localhost:5432/vsftpd
$ vsftpdmgr \
  -addr :8181 \
  -ca-file  /etc/ssl/certs/ca.crt \
  -cert-file /etc/ssl/certs/vsftpdmgr.crt \
  -key-file /etc/ssl/private/vsftpdmgr.key \
  /etc/vsftpd.passwd \
  /srv/ftp
```

vsftpdmgr tries to chmod and chown user local directories when the corresponding option is provided while updating an user, to avoid running the binary as a superuser it's recommended to restrict root privileges by changing the UNIX file capabilities and run the program as a normal user:
```
$ sudo setcap CAP_CHOWN,CAP_FOWNER=+ep vsftpdmgr
```

**WARNING**: for multi-server installation the pwdfile has to be accessible by all instances, e.g. put it on a nfs. Otherwise it can lead to unexpected behaviour.

## Systemd

```
[Unit]
Description=vsftpdmgr
After=network.target

[Service]
Type=simple
Environment=DATABASE_URL=postgres://user:pass@host:port/dbname
ExecStart=/usr/local/bin/vsftpdmgr -addr :8181 /srv/vsftpd/vsftpd.conf /srv/vsftpd.passwd

[Install]
WantedBy=multi-user.target
```

As well you may consider running the service under unprivileged user, but keep in mind that it won't be able to chown and chmod in some cases.

There're two options to protect your system:

1. Using systemd's options `ProtectSystem=`, `ProtectHome=`, etc.
1. Grant the service capabilities (TODO: figure out them).

## Testing

Don't forget to add `?sslmode=disable` if ssl is disabled for localhost connections (default behaviour).

```
$ export TEST_DATABASE_URL=postgres://user:pass@localhost:5432/vsftpd_test
$ make test
```
