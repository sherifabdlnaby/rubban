# Build Stage
FROM golang:1.13.7-alpine3.11 AS build-stage

LABEL app="build-rubban"
LABEL REPO="https://github.com/sherifabdlnaby/rubban"

ENV PROJPATH=/go/src/github.com/sherifabdlnaby/rubban

RUN apk add --no-cache git make

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

WORKDIR /go/src/github.com/sherifabdlnaby/rubban

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
ARG BUILD_DATE

LABEL   org.label-schema.build-date=$BUILD_DATE \
        org.label-schema.name="Rubban" \
        org.label-schema.description="Kibana Automatic Index Pattern Discovery and Other Curating Tasks." \
        org.label-schema.url="https://github.com/sherifabdlnaby/rubban" \
        org.label-schema.vcs-ref=$GIT_COMMIT \
        org.label-schema.vcs-url="https://github.com/sherifabdlnaby/rubban" \
        org.label-schema.version=$VERSION \
        org.label-schema.schema-version="1.0"

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:/opt/rubban/

WORKDIR /opt/rubban/

COPY ./rubban.yml.dist rubban.yml
COPY --from=build-stage /go/src/github.com/sherifabdlnaby/rubban/bin/rubban /opt/rubban/rubban
RUN chmod +x /opt/rubban/rubban

ENTRYPOINT ["/opt/rubban/rubban"]
