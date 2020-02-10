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
ARG GIT_DIRTY
ARG GIT_COMMIT_SHORT
ARG VERSION
ARG BUILD_DATE

LABEL   REPO="https://github.com/sherifabdlnaby/bosun" \
        GIT_COMMIT=$GIT_COMMIT \
        VERSION=$VERSION \
        org.label-schema.build-date=$BUILD_DATE \
        org.label-schema.name="Bosun" \
        org.label-schema.description="Kibana Automatic Index Pattern Discovery and Other Curating Tasks." \
        org.label-schema.url="https://github.com/sherifabdlnaby/bosun" \
        org.label-schema.vcs-ref=$GIT_COMMIT_SHORT \
        org.label-schema.vcs-url="https://github.com/sherifabdlnaby/bosun" \
        org.label-schema.version=$VERSION \
        org.label-schema.schema-version="1.0"

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:/opt/bosun/

WORKDIR /opt/bosun/

COPY ./bosun.yml.dist bosun.yml
COPY --from=build-stage /go/src/github.com/sherifabdlnaby/bosun/bin/bosun /opt/bosun/bosun
RUN chmod +x /opt/bosun/bosun

ENTRYPOINT ["/opt/bosun/bosun"]
