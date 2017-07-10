//+build skip

//go:generate openssl req -x509 -newkey rsa:2048 -keyout server.key -out server.crt -days 365 -nodes -subj "/C=US/ST=OR/L=Portland/O=Evil/OU=Corp/CN=localhost"
//go:generate openssl genrsa -out client.key 2048
//go:generate openssl req -new -key client.key -out client.csr -nodes -subj "/C=US/ST=IL/L=Chicago/O=Evil/OU=Corp/CN=localhost"
//go:generate openssl x509 -req -in client.csr -CA server.crt -CAkey server.key -CAcreateserial -out client.crt -days 365 -sha256

package testdata
