_build:
	mkdir -p _build

_build/http-probe: _build main.go
	go build -o _build/http-probe main.go

.PHONY: run
run: _build/http-probe
	_build/http-probe

.PHONY: clean
clean:
	rm -rf _build

http1: _build/http-probe
	PROTO_OPTION="http1" _build/http-probe

http2: _build/http-probe
	PROTO_OPTION="http2" _build/http-probe

h2c: _build/http-probe
	PROTO_OPTION="h2c" _build/http-probe
