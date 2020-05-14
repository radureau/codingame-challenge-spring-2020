all: build
	ls samples/ |sort -R|head -n1| sed 's/^/samples\//' | xargs cat | go run dist/response.go 

build:
ifeq (,$(shell which gocat))
	$(install_gocat)
endif
	@mkdir -p dist
	gocat -n -p main *.go > dist/response.go

define install_gocat
	GO111MODULE=off go get github.com/naegelejd/gocat
endef