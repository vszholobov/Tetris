module tetrisClient

go 1.23.0

toolchain go1.24.10

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.1
	github.com/jellydator/ttlcache/v3 v3.4.0
	github.com/mattn/go-tty v0.0.5
	github.com/vszholobov/tetrisLib v0.0.0
)

replace github.com/vszholobov/tetrisLib => ../lib

require (
	github.com/mattn/go-isatty v0.0.19 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
)
