# Build Stage
FROM golang:1.13 AS build-stage

LABEL app="build-bosun"
LABEL REPO="https://github.com/sherifabdlnaby/bosun"

ENV PROJPATH=/go/src/github.com/sherifabdlnaby/bosun

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

ADD . /go/src/github.com/sherifabdlnaby/bosun
WORKDIR /go/src/github.com/sherifabdlnaby/bosun

RUN make build-alpine

# Final Stage
FROM golang

ARG GIT_COMMIT
ARG VERSION
LABEL REPO="https://github.com/sherifabdlnaby/bosun"
LABEL GIT_COMMIT=$GIT_COMMIT
LABEL VERSION=$VERSION

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:/opt/bosun/bin

WORKDIR /opt/bosun/bin

COPY --from=build-stage /go/src/github.com/sherifabdlnaby/bosun/bin/bosun /opt/bosun/bin/
RUN chmod +x /opt/bosun/bin/bosun

# Create appuser
RUN adduser -D -g '' bosun
USER bosun

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["/opt/bosun/bin/bosun"]
