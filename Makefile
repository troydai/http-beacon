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