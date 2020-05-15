all: build
	ls samples/ |sort -R|head -n1| sed 's/^/samples\//' | xargs cat | go run dist/response.go 

build:
ifeq (,$(shell which gocat))
	$(install_gocat)
endif
	@mkdir -p dist
	gocat -n -p main *.go | sed -e s/__USER__/${USER}/ > dist/response.go
	

define install_gocat
	GO111MODULE=off go get github.com/naegelejd/gocat
endef

readASample:
	cp mapReader.go.backup dist/response.go
	@echo go to codingame and make sure it is synced

sample%: build
	cat samples/input$*.txt | go run dist/response.go

sample11: