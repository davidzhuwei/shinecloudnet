# Simple usage with a mounted data directory:
# > docker build -t scloud .
# > docker run -it -p 46657:46657 -p 46656:46656 -v ~/.scloud:/root/.scloud -v ~/.scloudcli:/root/.scloudcli scloud scloud init
# > docker run -it -p 46657:46657 -p 46656:46656 -v ~/.scloud:/root/.scloud -v ~/.scloudcli:/root/.scloudcli scloud scloud start
FROM golang:alpine AS build-env

# Set up dependencies
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python

# Set working directory for the build
WORKDIR /go/src/github.com/shinecloudfoundation/shinecloudnet

# Add source files
COPY . .

# Install minimum necessary dependencies, build Cosmos SDK, remove packages
RUN apk add --no-cache $PACKAGES && \
    make tools && \
    make install

# Final image
FROM alpine:edge

# Install ca-certificates
RUN apk add --update ca-certificates
WORKDIR /root

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/scloud /usr/bin/scloud
COPY --from=build-env /go/bin/scloudcli /usr/bin/scloudcli

# Run scloud by default, omit entrypoint to ease using container with scloudcli
CMD ["scloud"]
