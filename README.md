# vsftpdmgr [![CircleCI](https://circleci.com/gh/amenzhinsky/vsftpdmgr.svg?style=svg)](https://circleci.com/gh/amenzhinsky/vsftpdmgr)

[vsftpd](https://en.wikipedia.org/wiki/Vsftpd) users management daemon.

## API

Create new or update existing user:
```
$ curl localhost:8080/users/ -d '{"username": "test", "password": "test", "fs": {
	"mode": 0555,
	"owner": "ftp",
	"group": "ftp"
	"children": [
		{
			"name": "read"
			"mode": 0555,
			"owner": "ftp",
			"group": "ftp"
		},
		{
			"name": "write"
			"mode": 0755,
			"owner": "ftp",
			"group": "ftp"
		}
	]
}}'
```

Delete user:
```
$ curl -X DELETE localhost:8080/users/ -d '{"username": "test"}'
```

List all users:
```
$ curl localhost:8080/users/
```

## Running

The service requires a database storage, but currently only postgresql is supported.

`DATABASE_URL` is passed to the binary via environment variable instead of a flag for security reasons.

```
$ export DATABASE_URL=postgres://user:pass@localhost:5432/vsftpd
$ vsftpdmgr \
	-addr :8080 \
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
Environment=DATABASE_URL=postgres://user:pass@host:port/dbname
Type=simple
ExecStart=/usr/local/bin/vsftpdmgr \
	-addr :8080 \
	/srv/vsftpd \vsftpd.conf
	/srv/vsftpd.passwd

PrivateTmp=yes
PrivateDevices=yes
Restart=always
RestartSec=15
CapabilityBoundingSet=CAP_CHOWN CAP_FOWNER

[Install]
WantedBy=multi-user.target
```

## Testing

Don't forget to add `?sslmode=disable` if ssl is disabled for localhost connections (default behaviour).

```
$ export TEST_DATABASE_URL=postgres://user:pass@localhost:5432/vsftpd_test
$ make test
```
