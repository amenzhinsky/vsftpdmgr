# vsftpdmgr [![CircleCI](https://circleci.com/gh/amenzhinsky/vsftpdmgr.svg?style=svg)](https://circleci.com/gh/amenzhinsky/vsftpdmgr)

[vsftpd](https://en.wikipedia.org/wiki/Vsftpd) users management daemon.

## API

Create new or update existing user:
```
$ curl localhost:8080/users/ -d '{"username": "test", "password": "test"}'
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

Currently only postgresql is supported.

`DATABASE_URL` is passed to the binary via environment variable instead of a flag for security reasons.

```
$ export DATABASE_URL=postgres://user:pass@localhost:5432/vsftpd
$ vsftpdmgr /etc/vsftpd.passwd /srv/ftp
```

**WARNING**: for multi-server installation the pwdfile has to be accessible by all instances, e.g. put it on a nfs. Otherwise it can lead to unexpected behaviour.

## Testing

```
$ export TEST_DATABASE_URL=postgres://user:pass@localhost:5432/vsftpd_test?sslmode=disable
$ make test
```
