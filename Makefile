APP?=S3Adapter

GOOS?=linux
GOARCH?=amd64

VERSION?=$(shell git describe --tags --always)
NOW?=$(shell date -u '+%Y/%m/%d/%I:%M:%S%Z')
PROJECT?=github.com/solution9th/${APP}
CONTAINER_IMAGE?=solution9th/${APP}


LDFLAGS += -X "${PROJECT}/app.BuildTime=${NOW}"
LDFLAGS += -X "${PROJECT}/app.BuildVersion=${VERSION}"
LDFLAGS += -X "${PROJECT}/app.BuildAppName=${APP}"
BUILD_TAGS = ""
BUILD_FLAGS = "-v"

.PHONY: build build-local build-linux clean govet bindata docker-image

default: build


build: clean bindata govet
	CGO_ENABLED=0 GOOS= GOARCH= go build ${BUILD_FLAGS} -ldflags '${LDFLAGS}' -tags '${BUILD_TAGS}' -o ${APP} 

build-local: mock build

build-linux: clean mock bindata govet
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build ${BUILD_FLAGS} -ldflags '${LDFLAGS}' -tags '${BUILD_TAGS}' -o ${APP}


govet: 
	@ go vet . && go fmt ./... && \
	(if [[ "$(gofmt -d $(find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./tests/*" -not -path "./assets/*"))" == "" ]]; then echo "Good format"; else echo "Bad format"; exit 33; fi);

bindata: clean
	go get github.com/jteeuwen/go-bindata/...
	go-bindata -nomemcopy -pkg=conf \
		-debug=false \
		-o=conf/conf.go \
		conf/...

clean: 
	@ rm -fr ${APP} main conf/*.go

docker-image: 
	docker build -t $(CONTAINER_IMAGE):$(VERSION) -f ./Dockerfile .

mock:
	go get -u golang.org/x/tools/go/packages
	go get -u github.com/golang/mock/gomock
	go install github.com/golang/mock/mockgen
	mockgen ${PROJECT}/internal/db DB > mocks/mock_db/mock_db.go
	mockgen ${PROJECT}/internal/gateway S3Protocol,Gateway > mocks/mock_gateway/mock_gateway.go