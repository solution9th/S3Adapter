ARG GO_VERSION=1.12.6

FROM golang:${GO_VERSION}-alpine3.9 AS build-env

RUN apk --no-cache add build-base git

ENV GO111MODULE=off
WORKDIR ${GOPATH}/src/github.com/solution9th/S3Adapter

COPY . ${GOPATH}/src/github.com/solution9th/S3Adapter

RUN ls -alh && make build


FROM alpine:3.9

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

RUN apk update && apk add tzdata \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \ 
    && echo "Asia/Shanghai" > /etc/timezone

RUN apk add --update ca-certificates && rm -rf /var/cache/apk/*

COPY --from=build-env go/src/github.com/solution9th/S3Adapter/S3Adapter /main

EXPOSE 9091

ENTRYPOINT ["/main"]