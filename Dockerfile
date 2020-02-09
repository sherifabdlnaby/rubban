# Build Stage
FROM golang:1.13.7-alpine3.11 AS build-stage

LABEL app="build-bosun"
LABEL REPO="https://github.com/sherifabdlnaby/bosun"

ENV PROJPATH=/go/src/github.com/sherifabdlnaby/bosun

RUN apk add --no-cache git make

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

WORKDIR /go/src/github.com/sherifabdlnaby/bosun

COPY go.mod .
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download


COPY . .
RUN make build-alpine

# Final Stage
FROM alpine:3.11

ARG GIT_COMMIT
ARG VERSION
LABEL REPO="https://github.com/sherifabdlnaby/bosun"
LABEL GIT_COMMIT=$GIT_COMMIT
LABEL VERSION=$VERSION

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:/opt/bosun/

WORKDIR /opt/bosun/

COPY ./bosun.yml.dist bosun.yml
COPY --from=build-stage /go/src/github.com/sherifabdlnaby/bosun/bin/bosun /opt/bosun/bosun
RUN chmod +x /opt/bosun/bosun

ENTRYPOINT ["/opt/bosun/bosun"]
