# Stage 1 (to create a "build" image)
FROM golang:1.24 AS source

COPY . /go/src/github.com/payment-system/dq-vault/
WORKDIR /go/src/github.com/payment-system/dq-vault/

RUN go get -u github.com/Masterminds/glide && \
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build

# Stage 2 (to create a vault conatiner with executable)
FROM hashicorp/vault:1.15.6

# Make new directory for plugins
RUN mkdir /vault/plugins

RUN apk --no-cache add ca-certificates=* wget=* perl-utils=* && \
    wget -q -O /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub && \
    wget --progress=dot:giga https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.34-r0/glibc-2.34-r0.apk && \
    apk add --no-cache --force-overwrite glibc-2.34-r0.apk

# Copy executable from source to vault
COPY --from=source /go/src/github.com/payment-system/dq-vault/dq-vault /vault/plugins/dq-vault
RUN chown vault:vault /vault/plugins/dq-vault
