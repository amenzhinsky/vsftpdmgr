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

**WARNING**: for multi-server installation the pwdfile has to be accessible by all instances, e.g. put it on a nfs. Otherwise it can lead to unexpected behaviour.

## Testing

Don't forget to add `?sslmode=disable` if ssl is disabled for localhost connections (default behaviour).

```
$ export TEST_DATABASE_URL=postgres://user:pass@localhost:5432/vsftpd_test
$ make test
```
