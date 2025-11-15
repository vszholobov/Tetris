module tetrisClient

go 1.21

require (
	github.com/gorilla/websocket v1.5.1
	github.com/mattn/go-tty v0.0.5
	github.com/vszholobov/tetrisLib v0.0.0
)

replace github.com/vszholobov/tetrisLib => ../lib

require (
	github.com/mattn/go-isatty v0.0.10 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
)
